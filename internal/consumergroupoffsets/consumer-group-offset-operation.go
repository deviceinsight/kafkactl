package consumergroupoffsets

import (
	"strconv"

	"github.com/deviceinsight/kafkactl/internal/helpers"
	"golang.org/x/sync/errgroup"

	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/internal"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/pkg/errors"
)

type ResetConsumerGroupOffsetFlags struct {
	Topic             []string
	AllTopics         bool
	Partition         int32
	Offset            int64
	OldestOffset      bool
	NewestOffset      bool
	Execute           bool
	OutputFormat      string
	allowedGroupState string
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

	if (flags.Topic == nil || len(flags.Topic) == 0) && (!flags.AllTopics) {
		return errors.New("no topic specified")
	}

	if !flags.Execute {
		output.Warnf("nothing will be changed (include --execute to perform the reset)")
	}

	var (
		ctx          internal.ClientContext
		config       *sarama.Config
		err          error
		client       sarama.Client
		admin        sarama.ClusterAdmin
		descriptions []*sarama.GroupDescription
	)

	if ctx, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if config, err = internal.CreateClientConfig(&ctx); err != nil {
		return err
	}

	if client, err = internal.CreateClient(&ctx); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if admin, err = internal.CreateClusterAdmin(&ctx); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	var topics []string

	if flags.AllTopics {
		// retrieve all topics in the consumerGroup
		offsets, err := admin.ListConsumerGroupOffsets(groupName, nil)
		if err != nil {
			return errors.Wrap(err, "failed to list consumer group offsets")
		}
		for topic := range offsets.Blocks {
			topics = append(topics, topic)
		}
	} else {
		// verify that the provided topics exist
		existingTopics, err := client.Topics()
		if err != nil {
			return errors.Wrap(err, "failed to list available topics")
		}

		for _, topic := range flags.Topic {
			if !util.ContainsString(existingTopics, topic) {
				return errors.Errorf("topic does not exist: %s", topic)
			}
		}

		topics = flags.Topic
	}

	output.Debugf("reset consumer-group offset for topics: %v", topics)

	if flags.allowedGroupState == "" {
		// a reset is only allowed if group is empty (no one in the group)
		// for creation of the group state "Dead" is allowed
		flags.allowedGroupState = "Empty"
	}

	if flags.Execute {
		if descriptions, err = admin.DescribeConsumerGroups([]string{groupName}); err != nil {
			return errors.Wrap(err, "failed to describe consumer group")
		}

		for _, description := range descriptions {
			// https://stackoverflow.com/a/61745884/1115279
			if description.State != flags.allowedGroupState {
				return errors.Errorf("cannot reset offsets on consumer group %s. There are consumers assigned (state: %s)", groupName, description.State)
			}
		}
	}

	consumerGroup, err := sarama.NewConsumerGroup(ctx.Brokers, groupName, config)
	if err != nil {
		return errors.Errorf("failed to create consumer group %s: %v", groupName, err)
	}

	terminalCtx := helpers.CreateTerminalContext()

	consumeErrorGroup, _ := errgroup.WithContext(terminalCtx)
	consumeErrorGroup.SetLimit(100)

	for _, topic := range topics {
		topicName := topic
		consumeErrorGroup.Go(func() error {
			consumer := Consumer{
				client:    client,
				groupName: groupName,
				topicName: topicName,
				flags:     flags,
				ready:     make(chan bool),
			}

			err = consumerGroup.Consume(terminalCtx, []string{topicName}, &consumer)
			if err != nil {
				return err
			}
			<-consumer.ready
			return nil
		})
	}

	err = consumeErrorGroup.Wait()
	if err != nil {
		return err
	}

	err = consumerGroup.Close()
	if err != nil {
		return err
	}

	return nil
}

func (operation *ConsumerGroupOffsetOperation) CreateConsumerGroup(flags ResetConsumerGroupOffsetFlags, group string) error {

	flags.allowedGroupState = "Dead"
	flags.Execute = true
	flags.OutputFormat = "none"

	err := operation.ResetConsumerGroupOffset(flags, group)
	if err != nil {
		return err
	}
	output.Infof("consumer-group created: %s", group)
	return nil
}

type Consumer struct {
	ready     chan bool
	client    sarama.Client
	groupName string
	topicName string
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

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *Consumer) ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error {
	return nil
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

func getGroupOffset(offsetFetchResponse *sarama.OffsetFetchResponse, topic string, partition int32) int64 {
	block := offsetFetchResponse.Blocks[topic][partition]
	if block != nil {
		return block.Offset
	}
	return -1
}

type DeleteConsumerGroupOffsetFlags struct {
	Topic     string
	Partition int32
}

func (operation *ConsumerGroupOffsetOperation) DeleteConsumerGroupOffset(groupName string, flags DeleteConsumerGroupOffsetFlags) error {

	var (
		err        error
		context    internal.ClientContext
		admin      sarama.ClusterAdmin
		partitions []int32
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}
	defer admin.Close()

	offsets, err := admin.ListConsumerGroupOffsets(groupName, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to list group offsets: %s", groupName)
	}
	if _, ok := offsets.Blocks[flags.Topic]; !ok {
		return errors.Errorf("no offsets for topic: %s", flags.Topic)
	}
	if flags.Partition == -1 {
		// delete all existing offsets
		partitions = make([]int32, 0)
		for k, block := range offsets.Blocks[flags.Topic] {
			output.Infof("block : %s", block)
			partitions = append(partitions, int32(k))
		}
	} else {
		// check if the partition exists and delete only the offset for this partition
		if _, ok := offsets.Blocks[flags.Topic][flags.Partition]; !ok {
			return errors.Errorf("No offset for partition: %d", flags.Partition)
		}
		partitions = []int32{flags.Partition}
	}

	for _, partition := range partitions {
		if err = admin.DeleteConsumerGroupOffset(groupName, flags.Topic, partition); err != nil {
			return errors.Wrapf(err, "failed to delete consumer-group-offset [group: %s, topic: %s, partition: %d]",
				groupName, flags.Topic, flags.Partition)
		}
		output.Infof("consumer-group-offset deleted: [group: %s, topic: %s, partition: %d]",
			groupName, flags.Topic, partition)
	}
	return nil
}
