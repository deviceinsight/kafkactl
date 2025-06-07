package consume

import (
	"context"
	"math"
	"time"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/deviceinsight/kafkactl/v5/internal/util"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	ErrOffset = math.MinInt64
)

type PartitionConsumer struct {
	topic              string
	partitions         []int32
	client             *sarama.Client
	consumer           *sarama.Consumer
	partitionConsumers *errgroup.Group
}

func CreatePartitionConsumer(client *sarama.Client, topic string, partitions []int) (*PartitionConsumer, error) {

	consumer, err := sarama.NewConsumerFromClient(*client)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to start consumer: ")
	}

	var partitions2 []int32

	if len(partitions) == 0 {
		partitions2, err = consumer.Partitions(topic)

		if err != nil {
			return nil, errors.Wrap(err, "Failed to get the list of partitions")
		}
	} else {
		for _, partition := range partitions {
			partitions2 = append(partitions2, int32(partition))
		}
	}

	return &PartitionConsumer{
		topic:      topic,
		partitions: partitions2,
		client:     client,
		consumer:   &consumer,
	}, nil
}

func (c *PartitionConsumer) Start(ctx context.Context, flags Flags, messages chan<- *sarama.ConsumerMessage, stopConsumers <-chan bool) error {

	partitionErrorGroup, _ := errgroup.WithContext(ctx)

	var partitionContext context.Context
	c.partitionConsumers, partitionContext = errgroup.WithContext(ctx)

	for _, partition := range c.partitions {
		partitionID := partition
		partitionErrorGroup.Go(func() error {

			initialOffset, lastOffset, err := getOffsetBounds(c.client, c.topic, flags, partitionID)
			if err != nil {
				return err
			}
			pc, err := (*c.consumer).ConsumePartition(c.topic, partitionID, initialOffset)
			if err != nil {
				return errors.Errorf("Failed to start consumer for partition %d: %s", partitionID, err)
			}

			if lastOffset == -1 && (flags.Exit || flags.Tail > 0 || flags.ToTimestamp != "") {
				output.Debugf("Skipping empty partition %d", partitionID)
				return nil
			} else if lastOffset == -1 || initialOffset <= lastOffset {
				output.Debugf("Start consuming partition %d from offset %d to %d", partitionID, initialOffset, lastOffset)
			} else {
				output.Debugf("Skipping partition %d", partitionID)
				return nil
			}

			c.partitionConsumers.Go(func() error {

				messageChannel := pc.Messages()

			messageChannelRead:
				for {
					select {
					case message := <-messageChannel:
						if message != nil {
							messages <- message
							if lastOffset >= 0 && message.Offset >= lastOffset {
								output.Debugf("stop consuming partition %d limit offset reached: %d", partitionID, lastOffset)
								pc.AsyncClose()
								break messageChannelRead
							}
						}
					case <-time.After(5 * time.Second):
						if flags.Exit || flags.Tail > 0 {
							output.Warnf("timed-out while waiting for messages (https://github.com/deviceinsight/kafkactl/issues/67)")
							pc.AsyncClose()
							break messageChannelRead
						}
					case <-stopConsumers:
						output.Debugf("stop consumer on partition %d via channel", partitionID)
						pc.AsyncClose()
						break messageChannelRead
					case <-partitionContext.Done():
						output.Debugf("stop consumer on partition %d", partitionID)
						pc.AsyncClose()
						break messageChannelRead
					}
				}

				return nil
			})

			return nil
		})
	}

	return partitionErrorGroup.Wait()
}

func (c *PartitionConsumer) Wait() error {
	output.Debugf("waiting for partition consumers")
	return c.partitionConsumers.Wait()
}

func (c *PartitionConsumer) Close() error {
	output.Debugf("closing consumer")
	return (*c.consumer).Close()
}

func getOffsetBounds(client *sarama.Client, topic string, flags Flags, currentPartition int32) (int64, int64, error) {
	var startOffset int64
	var endOffset int64
	var err error
	if startOffset, err = getStartOffset(client, topic, flags, currentPartition); err != nil {
		return ErrOffset, ErrOffset, err
	}
	if endOffset, err = getEndOffset(client, topic, flags, currentPartition); err != nil {
		return ErrOffset, ErrOffset, err
	} else if startOffset == endOffset {
		endOffset = sarama.OffsetNewest //nothing to consume on this partition
	} else if endOffset != sarama.OffsetNewest {
		endOffset = endOffset - 1
	}
	if flags.Tail > 0 && startOffset == sarama.OffsetNewest {
		// When --tail is used compute startOffset so that it minimizes the number of messages consumed
		var oldestOffset int64
		if oldestOffset, err = (*client).GetOffset(topic, currentPartition, sarama.OffsetOldest); err != nil {
			return ErrOffset, ErrOffset, err
		}
		startOffset = max(oldestOffset, endOffset-int64(flags.Tail))
	}
	output.Debugf("consumer will consume offset %d to %d on partition %d", startOffset, endOffset, currentPartition)
	return startOffset, endOffset, nil
}

// Converts string to epoch unix timestamp
// The string might be null in that case, the flag is considered absent and the value -1 is returned
func ConvertToEpocUnixMillis(timestamp string) (int64, error) {
	if timestamp == "" {
		return -1, nil
	}
	aTime, err := util.ParseTimestamp(timestamp)
	if err != nil {
		return -1, err
	}
	return aTime.UnixMilli(), nil
}

func getStartOffset(client *sarama.Client, topic string, flags Flags, currentPartition int32) (int64, error) {
	var fromUnixMillis, err = ConvertToEpocUnixMillis(flags.FromTimestamp)
	if err != nil {
		return ErrOffset, err
	}
	if hasExclusiveConditions(fromUnixMillis > -1, flags.FromBeginning, len(flags.Offsets) > 0) {
		return ErrOffset, errors.Errorf("parameters '--from-timestamp', '--offset' and '--from-beginning' are exclusive")
	}
	if fromUnixMillis != -1 {
		return (*client).GetOffset(topic, currentPartition, fromUnixMillis)
	} else if flags.FromBeginning {
		return (*client).GetOffset(topic, currentPartition, sarama.OffsetOldest)
	} else if len(flags.Offsets) > 0 {
		return util.ExtractOffsetForPartition(flags.Offsets, currentPartition)
	}
	return sarama.OffsetNewest, nil
}

func getEndOffset(client *sarama.Client, topic string, flags Flags, currentPartition int32) (int64, error) {
	var toUnixMillis, err = ConvertToEpocUnixMillis(flags.ToTimestamp)
	if err != nil {
		return ErrOffset, err
	}
	if toUnixMillis > -1 {
		return (*client).GetOffset(topic, currentPartition, toUnixMillis)
	} else if flags.Exit || flags.Tail > 0 {
		var newestOffset int64
		var err error
		if newestOffset, err = (*client).GetOffset(topic, currentPartition, sarama.OffsetNewest); err != nil {
			return ErrOffset, err
		}
		return newestOffset, nil
	}
	return sarama.OffsetNewest, nil
}

func hasExclusiveConditions(flags ...bool) bool {
	value := 0
	for _, flag := range flags {
		if flag {
			value++
		}
	}
	return value > 1
}
