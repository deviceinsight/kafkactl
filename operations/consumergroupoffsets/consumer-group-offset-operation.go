package consumergroupoffsets

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/pkg/errors"
	"strconv"
)

type ResetConsumerGroupOffsetFlags struct {
	Topic        string
	Partition    int32
	Offset       int64
	OldestOffset bool
	NewestOffset bool
	Execute      bool
	OutputFormat string
}

type partitionOffsets struct {
	Partition     int32
	OldestOffset  int64 `json:"oldestOffset" yaml:"oldestOffset"`
	NewestOffset  int64 `json:"newestOffset" yaml:"newestOffset"`
	CurrentOffset int64 `json:"currentOffset" yaml:"currentOffset"`
	TargetOffset  int64 `json:"targetOffset" yaml:"targetOffset"`
}

type ConsumerGroupOffsetOperation struct {
}

func (operation *ConsumerGroupOffsetOperation) ResetConsumerGroupOffset(flags ResetConsumerGroupOffsetFlags, groupName string) error {

	if flags.Topic == "" {
		return errors.New("no topic specified")
	}

	if !flags.Execute {
		output.Warnf("nothing will be changed (include --execute to perform the reset)")
	}

	var (
		ctx          operations.ClientContext
		config       *sarama.Config
		err          error
		client       sarama.Client
		admin        sarama.ClusterAdmin
		descriptions []*sarama.GroupDescription
	)

	if ctx, err = operations.CreateClientContext(); err != nil {
		return err
	}

	if config, err = operations.CreateClientConfig(&ctx); err != nil {
		return err
	}

	if client, err = operations.CreateClient(&ctx); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if admin, err = operations.CreateClusterAdmin(&ctx); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	if topics, err := client.Topics(); err != nil {
		return errors.Wrap(err, "failed to list available topics")
	} else if !util.ContainsString(topics, flags.Topic) {
		return errors.Errorf("topic does not exist: %s", flags.Topic)
	}

	if flags.Execute {
		if descriptions, err = admin.DescribeConsumerGroups([]string{groupName}); err != nil {
			return errors.Wrap(err, "failed to describe consumer group")
		}

		for _, description := range descriptions {
			// https://stackoverflow.com/a/61745884/1115279
			if description.State != "Empty" {
				return errors.Errorf("cannot reset offsets on consumer group %s. There are consumers assigned (state: %s)", groupName, description.State)
			}
		}
	}

	consumerGroup, err := sarama.NewConsumerGroup(ctx.Brokers, groupName, config)
	if err != nil {
		return errors.Errorf("failed to create consumer group %s: %v", groupName, err)
	}

	backgroundCtx := context.Background()

	consumer := Consumer{
		client:    client,
		groupName: groupName,
		flags:     flags,
	}

	topics := []string{flags.Topic}

	consumer.ready = make(chan bool)
	err = consumerGroup.Consume(backgroundCtx, topics, &consumer)
	if err != nil {
		return err
	}

	<-consumer.ready

	err = consumerGroup.Close()
	if err != nil {
		return err
	}
	return nil
}

type Consumer struct {
	ready     chan bool
	client    sarama.Client
	groupName string
	flags     ResetConsumerGroupOffsetFlags
}

func (consumer *Consumer) Setup(session sarama.ConsumerGroupSession) error {

	flags := consumer.flags

	// admin.ListConsumerGroupOffsets(group, nil) can be used to fetch the offsets when
	// https://github.com/Shopify/sarama/pull/1374 is merged
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
		offset, err := resetOffset(consumer.client, flags.Partition, flags, groupOffsets, session)
		if err != nil {
			return err
		}
		offsets = append(offsets, offset)
	} else {

		partitions, err := consumer.client.Partitions(flags.Topic)
		if err != nil {
			return errors.Wrap(err, "failed to list partitions")
		}

		for _, partition := range partitions {
			offset, err := resetOffset(consumer.client, partition, flags, groupOffsets, session)
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
				strconv.FormatInt(int64(o.OldestOffset), 10), strconv.FormatInt(int64(o.NewestOffset), 10),
				strconv.FormatInt(int64(o.CurrentOffset), 10), strconv.FormatInt(int64(o.TargetOffset), 10)); err != nil {
				return err
			}
		}
		if err := tableWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *Consumer) ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error {
	return nil
}

func getPartitionOffsets(client sarama.Client, partition int32, flags ResetConsumerGroupOffsetFlags) (partitionOffsets, error) {

	var err error
	offsets := partitionOffsets{Partition: partition}

	if offsets.OldestOffset, err = client.GetOffset(flags.Topic, partition, sarama.OffsetOldest); err != nil {
		return offsets, errors.Errorf("failed to get offset for topic %s Partition %d: %v", flags.Topic, partition, err)
	}

	if offsets.NewestOffset, err = client.GetOffset(flags.Topic, partition, sarama.OffsetNewest); err != nil {
		return offsets, errors.Errorf("failed to get offset for topic %s Partition %d: %v", flags.Topic, partition, err)
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

func resetOffset(client sarama.Client, partition int32, flags ResetConsumerGroupOffsetFlags, groupOffsets *sarama.OffsetFetchResponse, session sarama.ConsumerGroupSession) (partitionOffsets, error) {
	offset, err := getPartitionOffsets(client, partition, flags)
	if err != nil {
		return offset, err
	}

	offset.CurrentOffset = getGroupOffset(groupOffsets, flags.Topic, partition)

	if flags.Execute {
		if offset.TargetOffset > offset.CurrentOffset {
			session.MarkOffset(flags.Topic, partition, offset.TargetOffset, "")
		} else if offset.TargetOffset < offset.CurrentOffset {
			session.ResetOffset(flags.Topic, partition, offset.TargetOffset, "")
		}
	}

	return offset, nil
}

func getGroupOffset(offsetFetchResponse *sarama.OffsetFetchResponse, topic string, partition int32) int64 {
	block := offsetFetchResponse.Blocks[topic][partition]
	if block != nil {
		return block.Offset
	} else {
		return -1
	}
}
