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
	"time"
)

type ConsumerFlags struct {
	PrintKeys       bool
	PrintTimestamps bool
	PrintAvroSchema bool
	PrintHeaders    bool
	OutputFormat    string
	Partitions      []int
	Offsets         []string
	FromBeginning   bool
	BufferSize      int
	Tail            int
	EncodeValue     string
	EncodeKey       string
}

type offset struct {
	relative bool
	start    int64
	diff     int64
}

type Interval struct {
	start offset
	end   offset
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
		signal.Notify(signals, os.Kill, os.Interrupt)
		<-signals
		output.Debugf("Initiating shutdown of consumer...")
		close(closing)
	}()

	output.Debugf("Start consuming topic: %s", topic)

	for _, partition := range partitions {
		wgPartition.Add(1)
		go func(partition int32) {
			defer wgPartition.Done()
			initialOffset, limit := getInitialOffset(client, topic, flags, partition)
			pc, err := c.ConsumePartition(topic, partition, initialOffset)
			if err != nil {
				output.Failf("Failed to start consumer for partition %d: %s", partition, err)
			}

			if limit > 0 {
				output.Debugf("Start consuming partition %d from offset %d", partition, initialOffset)
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
				messageCount := 0
				for message := range pc.Messages() {
					messages <- message
					messageCount += 1
					if limit > 0 && messageCount >= limit {
						output.Debugf("stop consuming partition %d limit reached: %d", partition, limit)
						pc.AsyncClose()
						break
					}
				}
			}(pc)
		}(partition)
	}

	if flags.Tail > 0 {

		wgPendingMessages.Add(1)

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

func getInitialOffset(client sarama.Client, topic string, flags ConsumerFlags, currentPartition int32) (int64, int) {

	if flags.Tail > 0 && len(flags.Offsets) > 0 {
		output.Failf("parameters offset and tail cannot be used together")
	} else if flags.Tail > 0 {

		var (
			err          error
			NewestOffset int64
			OldestOffset int64
		)

		if NewestOffset, err = client.GetOffset(topic, currentPartition, sarama.OffsetNewest); err != nil {
			output.Failf("failed to get offset for topic %s Partition %d: %v", topic, currentPartition, err)
		}

		if OldestOffset, err = client.GetOffset(topic, currentPartition, sarama.OffsetOldest); err != nil {
			output.Failf("failed to get offset for topic %s Partition %d: %v", topic, currentPartition, err)
		}

		offset := NewestOffset - int64(flags.Tail)
		limit := flags.Tail
		if offset < OldestOffset {
			offset = OldestOffset
			limit = int(NewestOffset - OldestOffset)
		}
		return offset, limit
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

			return offset, -1
		}
	}

	if flags.FromBeginning {
		return sarama.OffsetOldest, -1
	} else {
		return sarama.OffsetNewest, -1
	}
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
