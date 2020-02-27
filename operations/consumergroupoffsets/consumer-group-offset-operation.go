package consumergroupoffsets

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
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

func (operation *ConsumerGroupOffsetOperation) ResetConsumerGroupOffset(flags ResetConsumerGroupOffsetFlags, groupName string) {

	if flags.Topic == "" {
		output.Failf("no topic specified")
	}

	if !flags.Execute {
		output.Warnf("nothing will be changed (include --execute to perform the reset)")
	}

	ctx := operations.CreateClientContext()
	config := operations.CreateClientConfig(&ctx)

	var (
		err    error
		client sarama.Client
	)

	if client, err = operations.CreateClient(&ctx); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if topics, err := client.Topics(); err != nil {
		output.Failf("failed to list available topics: %v", err)
	} else if !util.ContainsString(topics, flags.Topic) {
		output.Failf("topic does not exist: %s", flags.Topic)
	}

	consumerGroup, err := sarama.NewConsumerGroup(ctx.Brokers, groupName, config)
	if err != nil {
		output.Failf("failed to create consumer group %s: %v", groupName, err)
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
		panic(err)
	}

	<-consumer.ready

	err = consumerGroup.Close()
	if err != nil {
		panic(err)
	}
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
		output.Failf("failed to get coordinator: %v", err)
	}

	request := &sarama.OffsetFetchRequest{
		// this will only work starting from version 0.10.2.0
		Version:       2,
		ConsumerGroup: consumer.groupName,
	}

	groupOffsets, err := coordinator.FetchOffset(request)
	if err != nil {
		output.Failf("failed to get fetch group offsets: %v", err)
	}

	offsets := make([]partitionOffsets, 0)

	if flags.Partition > -1 {
		offset := resetOffset(consumer.client, flags.Partition, flags, groupOffsets, session)
		offsets = append(offsets, offset)
	} else {

		partitions, err := consumer.client.Partitions(flags.Topic)
		if err != nil {
			output.Failf("failed to list partitions: %v", err)
		}

		for _, partition := range partitions {
			offset := resetOffset(consumer.client, partition, flags, groupOffsets, session)
			offsets = append(offsets, offset)
		}
	}

	if flags.OutputFormat != "" {
		output.PrintObject(offsets, flags.OutputFormat)
	} else {
		tableWriter := output.CreateTableWriter()
		tableWriter.WriteHeader("PARTITION", "OLDEST_OFFSET", "NEWEST_OFFSET", "CURRENT_OFFSET", "TARGET_OFFSET")
		for _, o := range offsets {
			tableWriter.Write(strconv.FormatInt(int64(o.Partition), 10),
				strconv.FormatInt(int64(o.OldestOffset), 10), strconv.FormatInt(int64(o.NewestOffset), 10),
				strconv.FormatInt(int64(o.CurrentOffset), 10), strconv.FormatInt(int64(o.TargetOffset), 10))
		}
		tableWriter.Flush()
	}
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	return nil
}

func getPartitionOffsets(client sarama.Client, partition int32, flags ResetConsumerGroupOffsetFlags) partitionOffsets {

	var err error
	offsets := partitionOffsets{Partition: partition}

	if offsets.OldestOffset, err = client.GetOffset(flags.Topic, partition, sarama.OffsetOldest); err != nil {
		output.Failf("failed to get offset for topic %s Partition %d: %v", flags.Topic, partition, err)
	}

	if offsets.NewestOffset, err = client.GetOffset(flags.Topic, partition, sarama.OffsetNewest); err != nil {
		output.Failf("failed to get offset for topic %s Partition %d: %v", flags.Topic, partition, err)
	}

	if flags.Offset > -1 {
		if flags.Offset < offsets.OldestOffset {
			output.Failf("cannot set offset for Partition %d: offset (%d) < oldest offset (%d)", partition, flags.Offset, offsets.OldestOffset)
		} else if flags.Offset > offsets.NewestOffset {
			output.Failf("cannot set offset for Partition %d: offset (%d) > newest offset (%d)", partition, flags.Offset, offsets.NewestOffset)
		} else {
			offsets.TargetOffset = flags.Offset
		}
	} else {
		if flags.OldestOffset {
			offsets.TargetOffset = offsets.OldestOffset
		} else if flags.NewestOffset {
			offsets.TargetOffset = offsets.NewestOffset
		} else {
			output.Failf("either offset,oldest,newest parameter needs to be specified")
		}
	}

	return offsets
}

func resetOffset(client sarama.Client, partition int32, flags ResetConsumerGroupOffsetFlags, groupOffsets *sarama.OffsetFetchResponse, session sarama.ConsumerGroupSession) partitionOffsets {
	offset := getPartitionOffsets(client, partition, flags)
	offset.CurrentOffset = getGroupOffset(groupOffsets, flags.Topic, partition)

	if flags.Execute {
		if offset.TargetOffset > offset.CurrentOffset {
			session.MarkOffset(flags.Topic, partition, offset.TargetOffset, "")
		} else if offset.TargetOffset < offset.CurrentOffset {
			session.ResetOffset(flags.Topic, partition, offset.TargetOffset, "")
		}
	}

	return offset
}

func getGroupOffset(offsetFetchResponse *sarama.OffsetFetchResponse, topic string, partition int32) int64 {
	block := offsetFetchResponse.Blocks[topic][partition]
	if block != nil {
		return block.Offset
	} else {
		return -1
	}
}
