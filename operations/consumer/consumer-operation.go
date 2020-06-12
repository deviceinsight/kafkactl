package consumer

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
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

func (operation *ConsumerOperation) Consume(topic string, flags ConsumerFlags) error {

	var (
		clientContext operations.ClientContext
		err           error
		client        sarama.Client
		topExists     bool
	)

	if clientContext, err = operations.CreateClientContext(); err != nil {
		return err
	}

	if client, err = operations.CreateClient(&clientContext); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if topExists, err = operations.TopicExists(&client, topic); err != nil {
		return errors.Wrap(err, "failed to read topics")
	}

	if !topExists {
		return errors.Errorf("topic '%s' does not exist", topic)
	}

	c, err := sarama.NewConsumer(clientContext.Brokers, client.Config())
	if err != nil {
		return errors.Wrap(err, "Failed to start consumer: ")
	}

	var deserializer MessageDeserializer

	if clientContext.AvroSchemaRegistry != "" {
		output.Debugf("using AvroMessageDeserializer")
		deserializer, err = CreateAvroMessageDeserializer(topic, clientContext.AvroSchemaRegistry)
		if err != nil {
			return err
		}
	} else {
		output.Debugf("using DefaultMessageDeserializer")
		deserializer = DefaultMessageDeserializer{}
	}

	var partitions []int32

	if flags.Partitions == nil || len(flags.Partitions) == 0 {
		partitions, err = c.Partitions(topic)

		if err != nil {
			return errors.Wrap(err, "Failed to get the list of partitions")
		}
	} else {
		for _, partition := range flags.Partitions {
			partitions = append(partitions, int32(partition))
		}
	}

	var (
		messages          = make(chan *sarama.ConsumerMessage, flags.BufferSize)
		errChannel        = make(chan error)
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
			initialOffset, lastOffset, err := getOffsetBounds(&client, topic, flags, partition)
			if err != nil {
				errChannel <- err
				return
			}
			pc, err := c.ConsumePartition(topic, partition, initialOffset)
			if err != nil {
				errChannel <- errors.Errorf("Failed to start consumer for partition %d: %s", partition, err)
				return
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
				err := deserializer.Deserialize(sortedMessages[lastIndex-i], flags)
				if err != nil {
					errChannel <- err
					return
				}
			}
		}()

	} else {
		//just print the messages
		go func() {
			defer wgPendingMessages.Done()
			for msg := range messages {
				err := deserializer.Deserialize(msg, flags)
				if err != nil {
					errChannel <- err
					return
				}
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

	select {
	case err := <-errChannel:
		return err
	default:
	}

	if err := c.Close(); err != nil {
		return errors.Wrap(err, "Failed to close consumer")
	}
	return nil
}

func getOffsetBounds(client *sarama.Client, topic string, flags ConsumerFlags, currentPartition int32) (int64, int64, error) {

	if flags.Exit && len(flags.Offsets) == 0 && !flags.FromBeginning {
		return -1, -1, errors.Errorf("parameter --exit has to be used in combination with --from-beginning or --offset")
	} else if flags.Tail > 0 && len(flags.Offsets) > 0 {
		return -1, -1, errors.Errorf("parameters --offset and --tail cannot be used together")
	} else if flags.Tail > 0 {

		newestOffset, oldestOffset, err := getBoundaryOffsets(client, topic, currentPartition)
		if err != nil {
			return -1, -1, err
		}

		minOffset := newestOffset - int64(flags.Tail)
		maxOffset := newestOffset - 1
		if minOffset < oldestOffset {
			minOffset = oldestOffset
		}
		return minOffset, maxOffset, nil
	}

	lastOffset := int64(-1)
	oldestOffset := sarama.OffsetOldest

	if flags.Exit {
		newestOffset, oldestOff, err := getBoundaryOffsets(client, topic, currentPartition)
		if err != nil {
			return -1, -1, err
		}
		lastOffset = newestOffset - 1
		oldestOffset = oldestOff
	}

	for _, offsetFlag := range flags.Offsets {
		offsetParts := strings.Split(offsetFlag, "=")

		if len(offsetParts) == 2 {

			partition, err := strconv.Atoi(offsetParts[0])
			if err != nil {
				return -1, -1, errors.Errorf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
			}

			if int32(partition) != currentPartition {
				continue
			}

			offset, err := strconv.ParseInt(offsetParts[1], 10, 64)
			if err != nil {
				return -1, -1, errors.Errorf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
			}

			return offset, lastOffset, nil
		}
	}

	if flags.FromBeginning {
		return oldestOffset, lastOffset, nil
	} else {
		return sarama.OffsetNewest, -1, nil
	}
}

func getBoundaryOffsets(client *sarama.Client, topic string, partition int32) (newestOffset int64, oldestOffset int64, err error) {

	if newestOffset, err = (*client).GetOffset(topic, partition, sarama.OffsetNewest); err != nil {
		return -1, -1, errors.Errorf("failed to get offset for topic %s Partition %d: %v", topic, partition, err)
	}

	if oldestOffset, err = (*client).GetOffset(topic, partition, sarama.OffsetOldest); err != nil {
		return -1, -1, errors.Errorf("failed to get offset for topic %s Partition %d: %v", topic, partition, err)
	}
	return newestOffset, oldestOffset, nil
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
