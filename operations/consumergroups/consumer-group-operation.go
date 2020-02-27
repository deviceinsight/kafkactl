package consumergroups

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
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
	ConsumerOffset int64 `json:"consumerOffset" yaml:"consumerOffset"`
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

func (operation *ConsumerGroupOperation) DescribeConsumerGroup(flags DescribeConsumerGroupFlags, group string) {

	ctx := operations.CreateClientContext()

	var (
		err          error
		client       sarama.Client
		admin        sarama.ClusterAdmin
		descriptions []*sarama.GroupDescription
	)

	if client, err = operations.CreateClient(&ctx); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if admin, err = operations.CreateClusterAdmin(&ctx); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	if descriptions, err = admin.DescribeConsumerGroups([]string{group}); err != nil {
		output.Failf("failed to describe consumer group: %v", err)
	}

	if flags.FilterTopic != "" {
		if topics, err := client.Topics(); err != nil {
			output.Failf("failed to list available topics: %v", err)
		} else if !util.ContainsString(topics, flags.FilterTopic) {
			output.Failf("topic does not exist: %s", flags.FilterTopic)
		}
	}

	offsets, err := admin.ListConsumerGroupOffsets(group, nil)

	if err != nil {
		output.Failf("failed to list consumer-group offsets: %v", err)
	}

	topicPartitions := createTopicPartitions(offsets, client, flags)

	for _, description := range descriptions {
		cg := consumerGroup{Name: description.GroupId, ProtocolType: description.ProtocolType}
		consumerGroupDescription := consumerGroupDescription{Group: cg, Protocol: description.Protocol, State: description.State, Topics: topicPartitions, Members: make([]consumerGroupMember, 0)}

		for _, member := range description.Members {

			memberAssignment, err := member.GetMemberAssignment()

			if err != nil {
				output.Failf("failed to get group member assignment: %v", err)
			}

			assignedPartitions := filterAssignedPartitions(memberAssignment.Topics, topicPartitions)

			consumerGroupDescription.Members = addMember(consumerGroupDescription.Members, member.ClientHost, member.ClientId, assignedPartitions)
		}

		sort.Slice(consumerGroupDescription.Members, func(i, j int) bool {
			return consumerGroupDescription.Members[i].ClientId < consumerGroupDescription.Members[j].ClientId
		})

		if !flags.PrintTopics {
			consumerGroupDescription.Topics = nil
		} else if flags.OutputFormat == "wide" || flags.OutputFormat == "" {
			tableWriter := output.CreateTableWriter()
			tableWriter.WriteHeader("TOPIC", "PARTITION", "NEWEST_OFFSET", "CONSUMER_OFFSET", "LAG")

			for _, topic := range consumerGroupDescription.Topics {
				for _, partition := range topic.Partitions {
					tableWriter.Write(topic.Name, strconv.Itoa(int(partition.Partition)), strconv.Itoa(int(partition.NewestOffset)), strconv.Itoa(int(partition.ConsumerOffset)), strconv.Itoa(int(partition.Lag)))
				}
			}

			tableWriter.Flush()
			output.PrintStrings("")
		}

		if !flags.PrintMembers {
			consumerGroupDescription.Members = nil
		} else if flags.OutputFormat == "wide" || flags.OutputFormat == "" {
			tableWriter := output.CreateTableWriter()
			tableWriter.WriteHeader("CLIENT_HOST", "CLIENT_ID", "TOPIC", "ASSIGNED_PARTITIONS")

			for _, m := range consumerGroupDescription.Members {
				for _, topic := range m.AssignedPartitions {
					partitions := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(topic.Partitions)), ","), "[]")
					tableWriter.Write(m.ClientHost, m.ClientId, topic.Name, partitions)
				}
			}

			tableWriter.Flush()
			output.PrintStrings("")
		}

		tableWriter := output.CreateTableWriter()

		if flags.OutputFormat == "" || flags.OutputFormat == "wide" {
			tableWriter.WriteHeader("PARTITION", "OLDEST_OFFSET", "NEWEST_OFFSET", "LEADER", "REPLICAS", "IN_SYNC_REPLICAS")
		} else if flags.OutputFormat != "json" && flags.OutputFormat != "yaml" {
			output.Failf("unknown outputFormat: %s", flags.OutputFormat)
		}

		if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
			output.PrintObject(consumerGroupDescription, flags.OutputFormat)
		}
	}
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

func createTopicPartitions(offsets *sarama.OffsetFetchResponse, client sarama.Client, flags DescribeConsumerGroupFlags) []topicPartitionOffsets {

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

			for partition := range offsets.Blocks[topic] {

				go func(partition int32) {

					offset, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
					if err != nil {
						output.Failf("failed to get offset for topic %s partition %d: %v", topic, partition, err)
					}

					lag := offset - offsets.Blocks[topic][partition].Offset

					if !flags.OnlyPartitionsWithLag || lag > 0 {
						partitionChannel <- partitionOffset{Partition: partition, NewestOffset: offset, ConsumerOffset: offsets.Blocks[topic][partition].Offset, Lag: lag}
					} else {
						partitionChannel <- partitionOffset{Partition: -1}
					}
				}(partition)
			}

			for range offsets.Blocks[topic] {
				partitionOffset := <-partitionChannel
				if partitionOffset.Partition != -1 {
					details = append(details, partitionOffset)
					totalLag += partitionOffset.Lag
				}
			}

			sort.Slice(details, func(i, j int) bool {
				return details[i].Partition < details[j].Partition
			})

			topicPartitions := topicPartitionOffsets{Name: topic, Partitions: details, TotalLag: totalLag}
			topicPartitionList = append(topicPartitionList, topicPartitions)
		}
	}

	return topicPartitionList
}

func (operation *ConsumerGroupOperation) GetConsumerGroups(flags GetConsumerGroupFlags) {

	ctx := operations.CreateClientContext()

	var (
		err    error
		admin  sarama.ClusterAdmin
		groups map[string]string
	)

	if admin, err = operations.CreateClusterAdmin(&ctx); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	// groups is a map from groupName to protocolType
	if groups, err = admin.ListConsumerGroups(); err != nil {
		output.Failf("failed to list consumer groups: %v", err)
	}

	groupNames := make([]string, 0, len(groups))
	for k := range groups {
		groupNames = append(groupNames, k)
	}

	if flags.FilterTopic != "" {
		groupNames = filterGroups(admin, groupNames, flags.FilterTopic)
	}

	sort.Strings(groupNames)

	tableWriter := output.CreateTableWriter()
	if flags.OutputFormat == "" {
		tableWriter.WriteHeader("CONSUMER_GROUP")
	} else if flags.OutputFormat == "compact" {
		output.PrintStrings(groupNames...)
		return
	} else if flags.OutputFormat == "wide" {
		tableWriter.WriteHeader("CONSUMER_GROUP", "PROTOCOL_TYPE")
	}

	for _, groupName := range groupNames {
		cg := consumerGroup{Name: groupName, ProtocolType: groups[groupName]}

		if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
			output.PrintObject(cg, flags.OutputFormat)
		} else if flags.OutputFormat == "wide" {
			tableWriter.Write(cg.Name, cg.ProtocolType)
		} else {
			tableWriter.Write(cg.Name)
		}
	}

	if flags.OutputFormat == "wide" || flags.OutputFormat == "" {
		tableWriter.Flush()
	}
}

func filterGroups(admin sarama.ClusterAdmin, groupNames []string, topic string) []string {

	var (
		err          error
		descriptions []*sarama.GroupDescription
	)

	if descriptions, err = admin.DescribeConsumerGroups(groupNames); err != nil {
		output.Failf("failed to describe consumer groups: %v", err)
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
				output.Failf("failed to get group member metadata: %v", err)
			}

			if util.ContainsString(metaData.Topics, topic) {
				if !util.ContainsString(topicGroups, description.GroupId) {
					topicGroups = append(topicGroups, description.GroupId)
				}
			}
		}
	}

	return topicGroups
}
