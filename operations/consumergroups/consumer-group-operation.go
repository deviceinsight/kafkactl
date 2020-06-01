package consumergroups

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"strings"
)

type consumerGroup struct {
	Name         string
	ProtocolType string
}

type topicPartitionOffsets struct {
	Name       string
	Partitions []partitionOffset `json:",omitempty" yaml:",omitempty"`
	TotalLag   int64             `json:"totalLag" yaml:"totalLag"`
}

type topicPartition struct {
	Name       string
	Partitions []int32 `json:"," yaml:",flow"`
}

type partitionOffset struct {
	Partition      int32
	NewestOffset   int64 `json:"newestOffset" yaml:"newestOffset"`
	OldestOffset   int64 `json:"oldestOffset" yaml:"oldestOffset"`
	ConsumerOffset int64 `json:"consumerOffset" yaml:"consumerOffset"`
	Lead           int64
	Lag            int64
}

type consumerGroupMember struct {
	ClientHost         string           `json:"clientHost" yaml:"clientHost"`
	ClientId           string           `json:"clientId" yaml:"clientId"`
	AssignedPartitions []topicPartition `json:"assignedPartitions" yaml:"assignedPartitions"`
}

type consumerGroupDescription struct {
	Group    consumerGroup
	Protocol string
	State    string
	Topics   []topicPartitionOffsets `json:",omitempty" yaml:",omitempty"`
	Members  []consumerGroupMember   `json:",omitempty" yaml:",omitempty"`
}

type DescribeConsumerGroupFlags struct {
	OnlyPartitionsWithLag bool
	FilterTopic           string
	OutputFormat          string
	PrintTopics           bool
	PrintMembers          bool
}

type GetConsumerGroupFlags struct {
	OutputFormat string
	FilterTopic  string
}

type ConsumerGroupOperation struct {
}

func (operation *ConsumerGroupOperation) DescribeConsumerGroup(flags DescribeConsumerGroupFlags, group string) error {

	var (
		err          error
		ctx          operations.ClientContext
		client       sarama.Client
		admin        sarama.ClusterAdmin
		descriptions []*sarama.GroupDescription
	)

	if ctx, err = operations.CreateClientContext(); err != nil {
		return err
	}

	if client, err = operations.CreateClient(&ctx); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if admin, err = operations.CreateClusterAdmin(&ctx); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	if descriptions, err = admin.DescribeConsumerGroups([]string{group}); err != nil {
		return errors.Wrap(err, "failed to describe consumer group")
	}

	if flags.FilterTopic != "" {
		if topics, err := client.Topics(); err != nil {
			return errors.Wrap(err, "failed to list available topics")
		} else if !util.ContainsString(topics, flags.FilterTopic) {
			return errors.Errorf("topic does not exist: %s", flags.FilterTopic)
		}
	}

	offsets, err := admin.ListConsumerGroupOffsets(group, nil)

	if err != nil {
		return errors.Wrap(err, "failed to list consumer-group offsets")
	}

	topicPartitions, err := createTopicPartitions(offsets, client, flags)
	if err != nil {
		return err
	}

	for _, description := range descriptions {
		cg := consumerGroup{Name: description.GroupId, ProtocolType: description.ProtocolType}
		consumerGroupDescription := consumerGroupDescription{Group: cg, Protocol: description.Protocol, State: description.State, Topics: topicPartitions, Members: make([]consumerGroupMember, 0)}

		for _, member := range description.Members {

			memberAssignment, err := member.GetMemberAssignment()

			var assignedPartitions map[string][]int32

			if err != nil {
				output.Warnf("failed to get group member assignment (%s, %s): %v", member.ClientHost, member.ClientId, err)
				assignedPartitions = make(map[string][]int32)
			} else {
				assignedPartitions = filterAssignedPartitions(memberAssignment.Topics, topicPartitions)
			}

			consumerGroupDescription.Members = addMember(consumerGroupDescription.Members, member.ClientHost, member.ClientId, assignedPartitions)
		}

		sort.Slice(consumerGroupDescription.Members, func(i, j int) bool {
			return consumerGroupDescription.Members[i].ClientId < consumerGroupDescription.Members[j].ClientId
		})

		if !flags.PrintTopics {
			consumerGroupDescription.Topics = nil
		} else if flags.OutputFormat == "wide" || flags.OutputFormat == "" {
			tableWriter := output.CreateTableWriter()

			if err := tableWriter.WriteHeader("TOPIC", "PARTITION", "NEWEST_OFFSET", "OLDEST_OFFSET", "CONSUMER_OFFSET", "LEAD", "LAG"); err != nil {
				return err
			}

			for _, topic := range consumerGroupDescription.Topics {
				for _, partition := range topic.Partitions {
					if err := tableWriter.Write(topic.Name, strconv.Itoa(int(partition.Partition)), strconv.Itoa(int(partition.NewestOffset)), strconv.Itoa(int(partition.OldestOffset)),
						strconv.Itoa(int(partition.ConsumerOffset)), strconv.Itoa(int(partition.Lead)), strconv.Itoa(int(partition.Lag))); err != nil {
						return err
					}
				}
			}

			if err := tableWriter.Flush(); err != nil {
				return err
			}
			output.PrintStrings("")
		}

		if !flags.PrintMembers {
			consumerGroupDescription.Members = nil
		} else if flags.OutputFormat == "wide" || flags.OutputFormat == "" {
			tableWriter := output.CreateTableWriter()
			if err := tableWriter.WriteHeader("CLIENT_HOST", "CLIENT_ID", "TOPIC", "ASSIGNED_PARTITIONS"); err != nil {
				return err
			}

			for _, m := range consumerGroupDescription.Members {
				for _, topic := range m.AssignedPartitions {
					partitions := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(topic.Partitions)), ","), "[]")
					if err := tableWriter.Write(m.ClientHost, m.ClientId, topic.Name, partitions); err != nil {
						return err
					}
				}
			}

			if err := tableWriter.Flush(); err != nil {
				return err
			}
			output.PrintStrings("")
		}

		tableWriter := output.CreateTableWriter()

		if flags.OutputFormat == "" || flags.OutputFormat == "wide" {
			if err := tableWriter.WriteHeader("PARTITION", "OLDEST_OFFSET", "NEWEST_OFFSET", "LEADER", "REPLICAS", "IN_SYNC_REPLICAS"); err != nil {
				return err
			}
		} else if flags.OutputFormat != "json" && flags.OutputFormat != "yaml" {
			return errors.Errorf("unknown outputFormat: %s", flags.OutputFormat)
		}

		if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
			if err := output.PrintObject(consumerGroupDescription, flags.OutputFormat); err != nil {
				return err
			}
		}
	}
	return nil
}

func filterAssignedPartitions(assignedPartitions map[string][]int32, topicPartitions []topicPartitionOffsets) map[string][]int32 {

	result := make(map[string][]int32)

	for topic, partitions := range assignedPartitions {
		for _, t := range topicPartitions {
			if t.Name == topic {
				result[topic] = partitions
			}
		}
	}

	return result
}

func addMember(members []consumerGroupMember, clientHost string, clientId string, assignedPartitions map[string][]int32) []consumerGroupMember {

	topicPartitionList := make([]topicPartition, 0)

	var topicsSorted []string

	for topic := range assignedPartitions {
		topicsSorted = append(topicsSorted, topic)
	}
	sort.Strings(topicsSorted)

	for _, topic := range topicsSorted {
		if topic != "" {
			topicPartitions := topicPartition{Name: topic, Partitions: assignedPartitions[topic]}
			topicPartitionList = append(topicPartitionList, topicPartitions)
		}
	}

	if len(assignedPartitions) == 0 {
		return members
	} else {
		member := consumerGroupMember{ClientHost: clientHost, ClientId: clientId, AssignedPartitions: topicPartitionList}
		return append(members, member)
	}
}

func createTopicPartitions(offsets *sarama.OffsetFetchResponse, client sarama.Client, flags DescribeConsumerGroupFlags) ([]topicPartitionOffsets, error) {

	topicPartitionList := make([]topicPartitionOffsets, 0)

	var topicsSorted []string

	for topic := range offsets.Blocks {
		if flags.FilterTopic == "" || topic == flags.FilterTopic {
			topicsSorted = append(topicsSorted, topic)
		}
	}
	sort.Strings(topicsSorted)

	for _, topic := range topicsSorted {
		if topic != "" {

			details := make([]partitionOffset, 0, len(offsets.Blocks[topic]))

			var totalLag int64 = 0

			partitionChannel := make(chan partitionOffset)
			errChannel := make(chan error)

			for partition := range offsets.Blocks[topic] {

				go func(partition int32) {

					newestOffset, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
					if err != nil {
						errChannel <- errors.Errorf("failed to get newest offset for topic %s partition %d: %v", topic, partition, err)
						return
					}

					oldestOffset, err := client.GetOffset(topic, partition, sarama.OffsetOldest)
					if err != nil {
						errChannel <- errors.Errorf("failed to get oldest offset for topic %s partition %d: %v", topic, partition, err)
						return
					}

					lead := offsets.Blocks[topic][partition].Offset - oldestOffset
					lag := newestOffset - offsets.Blocks[topic][partition].Offset

					if !flags.OnlyPartitionsWithLag || lag > 0 {
						partitionChannel <- partitionOffset{Partition: partition, NewestOffset: newestOffset, OldestOffset: oldestOffset,
							ConsumerOffset: offsets.Blocks[topic][partition].Offset, Lead: lead, Lag: lag}
					} else {
						partitionChannel <- partitionOffset{Partition: -1}
					}
				}(partition)
			}

			for range offsets.Blocks[topic] {
				select {
				case partitionOffset := <-partitionChannel:
					if partitionOffset.Partition != -1 {
						details = append(details, partitionOffset)
						totalLag += partitionOffset.Lag
					}
				case err := <-errChannel:
					return nil, err
				}
			}

			sort.Slice(details, func(i, j int) bool {
				return details[i].Partition < details[j].Partition
			})

			topicPartitions := topicPartitionOffsets{Name: topic, Partitions: details, TotalLag: totalLag}
			topicPartitionList = append(topicPartitionList, topicPartitions)
		}
	}

	return topicPartitionList, nil
}

func (operation *ConsumerGroupOperation) GetConsumerGroups(flags GetConsumerGroupFlags) error {

	var (
		ctx    operations.ClientContext
		err    error
		admin  sarama.ClusterAdmin
		groups map[string]string
	)

	if ctx, err = operations.CreateClientContext(); err != nil {
		return err
	}

	if admin, err = operations.CreateClusterAdmin(&ctx); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	// groups is a map from groupName to protocolType
	if groups, err = admin.ListConsumerGroups(); err != nil {
		return errors.Wrap(err, "failed to list consumer groups")
	}

	groupNames := make([]string, 0, len(groups))
	for k := range groups {
		groupNames = append(groupNames, k)
	}

	if flags.FilterTopic != "" {
		groupNames, err = filterGroups(admin, groupNames, flags.FilterTopic)
		if err != nil {
			return err
		}
	}

	sort.Strings(groupNames)

	tableWriter := output.CreateTableWriter()
	if flags.OutputFormat == "" {
		if err := tableWriter.WriteHeader("CONSUMER_GROUP"); err != nil {
			return err
		}
	} else if flags.OutputFormat == "compact" {
		output.PrintStrings(groupNames...)
		return nil
	} else if flags.OutputFormat == "wide" {
		if err := tableWriter.WriteHeader("CONSUMER_GROUP", "PROTOCOL_TYPE"); err != nil {
			return err
		}
	}

	for _, groupName := range groupNames {
		cg := consumerGroup{Name: groupName, ProtocolType: groups[groupName]}

		if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
			if err := output.PrintObject(cg, flags.OutputFormat); err != nil {
				return err
			}
		} else if flags.OutputFormat == "wide" {
			if err := tableWriter.Write(cg.Name, cg.ProtocolType); err != nil {
				return err
			}
		} else {
			if err := tableWriter.Write(cg.Name); err != nil {
				return err
			}
		}
	}

	if flags.OutputFormat == "wide" || flags.OutputFormat == "" {
		if err := tableWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func filterGroups(admin sarama.ClusterAdmin, groupNames []string, topic string) ([]string, error) {

	var (
		err          error
		descriptions []*sarama.GroupDescription
	)

	if descriptions, err = admin.DescribeConsumerGroups(groupNames); err != nil {
		return nil, errors.Wrap(err, "failed to describe consumer groups")
	}

	topicGroups := make([]string, 0)

	for _, description := range descriptions {

		if description.ProtocolType != "consumer" {
			output.Debugf("do not filter on group %s, because protocolType is: %s", description.GroupId, description.ProtocolType)
			continue
		}

		for _, member := range description.Members {

			metaData, err := member.GetMemberMetadata()

			if err != nil {
				return nil, errors.Wrap(err, "failed to get group member metadata")
			}

			if util.ContainsString(metaData.Topics, topic) {
				if !util.ContainsString(topicGroups, description.GroupId) {
					topicGroups = append(topicGroups, description.GroupId)
				}
			}
		}
	}

	return topicGroups, nil
}
