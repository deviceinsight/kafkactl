package consumergroupoffsets

import (
	"strconv"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
)

type partitionOffsets struct {
	Partition     int32
	OldestOffset  int64 `json:"oldestOffset" yaml:"oldestOffset"`
	NewestOffset  int64 `json:"newestOffset" yaml:"newestOffset"`
	CurrentOffset int64 `json:"currentOffset" yaml:"currentOffset"`
	TargetOffset  int64 `json:"targetOffset" yaml:"targetOffset"`
}

type OffsetResettingConsumer struct {
	ready     chan bool
	client    sarama.Client
	groupName string
	topicName string
	flags     ResetConsumerGroupOffsetFlags
}

func (consumer *OffsetResettingConsumer) Setup(session sarama.ConsumerGroupSession) error {

	flags := consumer.flags

	// admin.ListConsumerGroupOffsets(group, nil) can be used to fetch the offsets when
	// https://github.com/IBM/sarama/pull/1374 is merged
	coordinator, err := consumer.client.Coordinator(consumer.groupName)
	if err != nil {
		return errors.Wrap(err, "failed to get coordinator")
	}

	request := &sarama.OffsetFetchRequest{
		// this will only work starting from version 0.10.2.0
		Version:       2,
		ConsumerGroup: consumer.groupName,
	}

	groupOffsets, err := coordinator.FetchOffset(request)
	if err != nil {
		return errors.Wrap(err, "failed to get fetch group offsets")
	}

	offsets := make([]partitionOffsets, 0)

	if flags.Partition > -1 {
		offset, err := resetOffset(consumer.client, consumer.topicName, flags.Partition, flags, groupOffsets, session)
		if err != nil {
			return err
		}
		offsets = append(offsets, offset)
	} else {

		partitions, err := consumer.client.Partitions(consumer.topicName)
		if err != nil {
			return errors.Wrap(err, "failed to list partitions")
		}

		for _, partition := range partitions {
			offset, err := resetOffset(consumer.client, consumer.topicName, partition, flags, groupOffsets, session)
			if err != nil {
				return err
			}
			offsets = append(offsets, offset)
		}
	}

	if flags.OutputFormat != "" {
		if err := output.PrintObject(offsets, flags.OutputFormat); err != nil {
			return err
		}
	} else {
		tableWriter := output.CreateTableWriter()
		if err := tableWriter.WriteHeader("PARTITION", "OLDEST_OFFSET", "NEWEST_OFFSET", "CURRENT_OFFSET", "TARGET_OFFSET"); err != nil {
			return err
		}
		for _, o := range offsets {
			if err := tableWriter.Write(strconv.FormatInt(int64(o.Partition), 10),
				strconv.FormatInt(o.OldestOffset, 10), strconv.FormatInt(o.NewestOffset, 10),
				strconv.FormatInt(o.CurrentOffset, 10), strconv.FormatInt(o.TargetOffset, 10)); err != nil {
				return err
			}
		}
		if err := tableWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (consumer *OffsetResettingConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *OffsetResettingConsumer) ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error {
	return nil
}

func resetOffset(client sarama.Client, topic string, partition int32, flags ResetConsumerGroupOffsetFlags, groupOffsets *sarama.OffsetFetchResponse, session sarama.ConsumerGroupSession) (partitionOffsets, error) {
	offset, err := getPartitionOffsets(client, topic, partition, flags)
	if err != nil {
		return offset, err
	}

	offset.CurrentOffset = getGroupOffset(groupOffsets, topic, partition)

	if flags.Execute {
		if offset.TargetOffset > offset.CurrentOffset {
			session.MarkOffset(topic, partition, offset.TargetOffset, "")
		} else if offset.TargetOffset < offset.CurrentOffset {
			session.ResetOffset(topic, partition, offset.TargetOffset, "")
		}
	}

	return offset, nil
}

func getPartitionOffsets(client sarama.Client, topic string, partition int32, flags ResetConsumerGroupOffsetFlags) (partitionOffsets, error) {

	var err error
	offsets := partitionOffsets{Partition: partition}

	if offsets.OldestOffset, err = client.GetOffset(topic, partition, sarama.OffsetOldest); err != nil {
		return offsets, errors.Errorf("failed to get offset for topic %s Partition %d: %v", topic, partition, err)
	}

	if offsets.NewestOffset, err = client.GetOffset(topic, partition, sarama.OffsetNewest); err != nil {
		return offsets, errors.Errorf("failed to get offset for topic %s Partition %d: %v", topic, partition, err)
	}

	if flags.Offset > -1 {
		if flags.Offset < offsets.OldestOffset {
			return offsets, errors.Errorf("cannot set offset for Partition %d: offset (%d) < oldest offset (%d)", partition, flags.Offset, offsets.OldestOffset)
		} else if flags.Offset > offsets.NewestOffset {
			return offsets, errors.Errorf("cannot set offset for Partition %d: offset (%d) > newest offset (%d)", partition, flags.Offset, offsets.NewestOffset)
		} else {
			offsets.TargetOffset = flags.Offset
		}
	} else {
		if flags.OldestOffset {
			offsets.TargetOffset = offsets.OldestOffset
		} else if flags.NewestOffset {
			offsets.TargetOffset = offsets.NewestOffset
		} else {
			return offsets, errors.New("either offset,oldest,newest parameter needs to be specified")
		}
	}

	return offsets, nil
}

func getGroupOffset(offsetFetchResponse *sarama.OffsetFetchResponse, topic string, partition int32) int64 {
	block := offsetFetchResponse.Blocks[topic][partition]
	if block != nil {
		return block.Offset
	}
	return -1
}
