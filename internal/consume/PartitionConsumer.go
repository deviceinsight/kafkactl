package consume

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
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

			if lastOffset == -1 && (flags.Exit || flags.Tail > 0) {
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

	if flags.Exit && flags.FromTs != -1 {

		var newestOffset int64
		var oldestOffset int64
		var err error
		if oldestOffset, err = (*client).GetOffset(topic, currentPartition, flags.FromTs); err != nil {
			return -1, -1, errors.Errorf("failed to get offset for topic %s Partition %d: %v", topic, currentPartition, err)
		}

		if newestOffset, err = (*client).GetOffset(topic, currentPartition, flags.EndTs); err != nil {
			return -1, -1, errors.Errorf("failed to get offset for topic %s Partition %d: %v", topic, currentPartition, err)
		}

		return oldestOffset, newestOffset, err

	} else if flags.Exit && len(flags.Offsets) == 0 && !flags.FromBeginning {
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
	}
	return sarama.OffsetNewest, -1, nil
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
