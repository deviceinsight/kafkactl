package consumer

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ConsumerFlags struct {
	PrintKeys       bool
	PrintTimestamps bool
	PrintAvroSchema bool
	PrintHeaders    bool
	OutputFormat    string
	Separator       string
	Partitions      []int
	Offsets         []string
	FromBeginning   bool
	BufferSize      int
	Tail            int
	Exit            bool
	EncodeValue     string
	EncodeKey       string
}

type ConsumedMessage struct {
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Timestamp *time.Time
}

type ConsumerOperation struct {
}

func (operation *ConsumerOperation) Consume(topic string, flags ConsumerFlags) {

	clientContext := operations.CreateClientContext()

	var (
		err       error
		client    sarama.Client
		topExists bool
	)

	if client, err = operations.CreateClient(&clientContext); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if topExists, err = operations.TopicExists(&client, topic); err != nil {
		output.Failf("failed to read topics err=%v", err)
	}

	if !topExists {
		output.Failf("topic '%s' does not exist", topic)
	}

	c, err := sarama.NewConsumer(clientContext.Brokers, client.Config())
	if err != nil {
		output.Failf("Failed to start consumer: %s", err)
	}

	var deserializer MessageDeserializer

	if clientContext.AvroSchemaRegistry != "" {
		output.Debugf("using AvroMessageDeserializer")
		deserializer = CreateAvroMessageDeserializer(topic, clientContext.AvroSchemaRegistry)
	} else {
		output.Debugf("using DefaultMessageDeserializer")
		deserializer = DefaultMessageDeserializer{}
	}

	var partitions []int32

	if flags.Partitions == nil || len(flags.Partitions) == 0 {
		partitions, err = c.Partitions(topic)

		if err != nil {
			output.Failf("Failed to get the list of partitions: %s", err)
		}
	} else {
		for _, partition := range flags.Partitions {
			partitions = append(partitions, int32(partition))
		}
	}

	var (
		messages          = make(chan *sarama.ConsumerMessage, flags.BufferSize)
		closing           = make(chan struct{})
		wgPartition       sync.WaitGroup
		wgConsumerActive  sync.WaitGroup
		wgPendingMessages sync.WaitGroup
	)

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
		<-signals
		output.Debugf("Initiating shutdown of consumer...")
		close(closing)
	}()

	output.Debugf("Start consuming topic: %s", topic)

	for _, partition := range partitions {
		wgPartition.Add(1)
		go func(partition int32) {
			defer wgPartition.Done()
			initialOffset, lastOffset := getOffsetBounds(&client, topic, flags, partition)
			pc, err := c.ConsumePartition(topic, partition, initialOffset)
			if err != nil {
				output.Failf("Failed to start consumer for partition %d: %s", partition, err)
			}

			if lastOffset == -1 || initialOffset <= lastOffset {
				output.Debugf("Start consuming partition %d from offset %d to %d", partition, initialOffset, lastOffset)
			} else {
				output.Debugf("Skipping partition %d", partition)
				return
			}

			go func(pc sarama.PartitionConsumer) {
				<-closing
				pc.AsyncClose()
			}(pc)

			wgConsumerActive.Add(1)
			go func(pc sarama.PartitionConsumer) {
				defer wgConsumerActive.Done()
				for message := range pc.Messages() {
					messages <- message
					if lastOffset >= 0 && message.Offset >= lastOffset {
						output.Debugf("stop consuming partition %d limit reached: %d", partition, lastOffset)
						pc.AsyncClose()
						break
					}
				}
			}(pc)
		}(partition)
	}

	wgPendingMessages.Add(1)

	if flags.Tail > 0 {
		go func() {
			defer wgPendingMessages.Done()

			sortedMessages := make([]*sarama.ConsumerMessage, 0)

			for msg := range messages {
				sortedMessages = insertSorted(sortedMessages, msg)
				if len(sortedMessages) > flags.Tail {
					sortedMessages = sortedMessages[:flags.Tail]
				}
			}
			lastIndex := len(sortedMessages) - 1
			for i := range sortedMessages {
				deserializer.Deserialize(sortedMessages[lastIndex-i], flags)
			}
		}()

	} else {
		//just print the messages
		go func() {
			defer wgPendingMessages.Done()
			for msg := range messages {
				deserializer.Deserialize(msg, flags)
			}
		}()
	}

	wgPartition.Wait()
	output.Debugf("Done waiting for partitions")
	wgConsumerActive.Wait()
	output.Debugf("Done consuming topic: %s", topic)
	close(messages)
	wgPendingMessages.Wait()
	output.Debugf("Done waiting for messages")

	if err := c.Close(); err != nil {
		output.Failf("Failed to close consumer: ", err)
	}
}

func getOffsetBounds(client *sarama.Client, topic string, flags ConsumerFlags, currentPartition int32) (int64, int64) {

	if flags.Exit && len(flags.Offsets) == 0 && !flags.FromBeginning {
		output.Failf("parameter --exit has to be used in combination with --from-beginning or --offset")
	} else if flags.Tail > 0 && len(flags.Offsets) > 0 {
		output.Failf("parameters --offset and --tail cannot be used together")
	} else if flags.Tail > 0 {

		newestOffset, oldestOffset := getBoundaryOffsets(client, topic, currentPartition)

		minOffset := newestOffset - int64(flags.Tail)
		maxOffset := newestOffset - 1
		if minOffset < oldestOffset {
			minOffset = oldestOffset
		}
		return minOffset, maxOffset
	}

	lastOffset := int64(-1)
	oldestOffset := sarama.OffsetOldest

	if flags.Exit {
		newestOffset, oldestOff := getBoundaryOffsets(client, topic, currentPartition)
		lastOffset = newestOffset - 1
		oldestOffset = oldestOff
	}

	for _, offsetFlag := range flags.Offsets {
		offsetParts := strings.Split(offsetFlag, "=")

		if len(offsetParts) == 2 {

			partition, err := strconv.Atoi(offsetParts[0])
			if err != nil {
				output.Failf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
			}

			if int32(partition) != currentPartition {
				continue
			}

			offset, err := strconv.ParseInt(offsetParts[1], 10, 64)
			if err != nil {
				output.Failf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
			}

			return offset, lastOffset
		}
	}

	if flags.FromBeginning {
		return oldestOffset, lastOffset
	} else {
		return sarama.OffsetNewest, -1
	}
}

func getBoundaryOffsets(client *sarama.Client, topic string, partition int32) (newestOffset int64, oldestOffset int64) {
	var err error

	if newestOffset, err = (*client).GetOffset(topic, partition, sarama.OffsetNewest); err != nil {
		output.Failf("failed to get offset for topic %s Partition %d: %v", topic, partition, err)
	}

	if oldestOffset, err = (*client).GetOffset(topic, partition, sarama.OffsetOldest); err != nil {
		output.Failf("failed to get offset for topic %s Partition %d: %v", topic, partition, err)
	}
	return newestOffset, oldestOffset
}

func insertSorted(messages []*sarama.ConsumerMessage, message *sarama.ConsumerMessage) []*sarama.ConsumerMessage {
	index := sort.Search(len(messages), func(i int) bool {
		return messages[i].Timestamp.Before(message.Timestamp)
	})
	messages = append(messages, nil)
	copy(messages[index+1:], messages[index:])
	messages[index] = message
	return messages
}
