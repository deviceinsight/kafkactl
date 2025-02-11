package consumergroupoffsets

import (
	"github.com/deviceinsight/kafkactl/v5/internal/helpers"
	"golang.org/x/sync/errgroup"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/deviceinsight/kafkactl/v5/internal/util"
	"github.com/pkg/errors"
)

type ResetConsumerGroupOffsetFlags struct {
	Topic        []string
	AllTopics    bool
	Partition    int32
	Offset       int64
	OldestOffset bool
	NewestOffset bool
	Execute      bool
	OutputFormat string
	ToDatetime   string
}

type ConsumerGroupOffsetOperation struct {
}

func (operation *ConsumerGroupOffsetOperation) ResetConsumerGroupOffset(flags ResetConsumerGroupOffsetFlags, groupName string) error {

	if (len(flags.Topic) == 0) && (!flags.AllTopics) {
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

	if flags.Execute {
		if descriptions, err = admin.DescribeConsumerGroups([]string{groupName}); err != nil {
			return errors.Wrap(err, "failed to describe consumer group")
		}

		for _, description := range descriptions {
			// https://stackoverflow.com/a/61745884
			if description.State != "Empty" && description.State != "Dead" {
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
			consumer := OffsetResettingConsumer{
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

	flags.Execute = true
	flags.OutputFormat = "none"

	err := operation.ResetConsumerGroupOffset(flags, group)
	if err != nil {
		return err
	}
	output.Infof("consumer-group created: %s", group)
	return nil
}

type PartitionOffset struct {
	Offset   int64
	Metadata string
}

func (operation *ConsumerGroupOffsetOperation) CloneConsumerGroup(srcGroup, targetGroup string) error {
	var (
		err                       error
		context                   internal.ClientContext
		config                    *sarama.Config
		admin                     sarama.ClusterAdmin
		srcOffsets, targetOffsets *sarama.OffsetFetchResponse
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if config, err = internal.CreateClientConfig(&context); err != nil {
		return err
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	if srcOffsets, err = admin.ListConsumerGroupOffsets(srcGroup, nil); err != nil {
		return errors.Wrapf(err, "failed to get consumerGroup '%s' offsets", srcGroup)
	}

	if len(srcOffsets.Blocks) == 0 {
		return errors.Errorf("consumerGroup '%s' does not contain offsets", srcGroup)
	}

	if targetOffsets, err = admin.ListConsumerGroupOffsets(targetGroup, nil); err != nil {
		return errors.Wrapf(err, "failed to get consumerGroup '%s' offsets", targetGroup)
	}

	if len(targetOffsets.Blocks) != 0 {
		return errors.Errorf("consumerGroup '%s' contains offsets", targetGroup)
	}

	topicPartitionOffsets := make(map[string]map[int32]PartitionOffset) // topic->partition->offset

	for topic, partitions := range srcOffsets.Blocks {
		p := topicPartitionOffsets[topic]
		if p == nil {
			p = make(map[int32]PartitionOffset)
		}

		for partition, block := range partitions {
			p[partition] = PartitionOffset{Offset: block.Offset, Metadata: block.Metadata}
		}

		topicPartitionOffsets[topic] = p
	}

	consumerGroup, err := sarama.NewConsumerGroup(context.Brokers, targetGroup, config)
	if err != nil {
		return errors.Errorf("failed to create consumer group %s: %v", targetGroup, err)
	}

	terminalCtx := helpers.CreateTerminalContext()

	consumeErrorGroup, _ := errgroup.WithContext(terminalCtx)
	consumeErrorGroup.SetLimit(100)

	for topic, partitionOffsets := range topicPartitionOffsets {
		topicName, offsets := topic, partitionOffsets
		consumeErrorGroup.Go(func() error {
			consumer := OffsetSettingConsumer{
				Topic:            topicName,
				PartitionOffsets: offsets,
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

	output.Infof("consumer-group %s cloned to %s", srcGroup, targetGroup)

	return nil
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
