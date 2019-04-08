package consumergroups

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"sort"
)

type consumerGroup struct {
	Name         string
	ProtocolType string
}

type topicPartitions struct {
	Topic            string
	Partitions       []int32           `json:",omitempty" yaml:",omitempty,flow"`
	PartitionDetails []partitionDetail `json:",omitempty" yaml:",omitempty"`
	TotalLag         int64
}

type partitionDetail struct {
	Partition      int32
	NewestOffset   int64
	ConsumerOffset int64
	Lag            int64
}

type consumerGroupMember struct {
	Host               string
	ClientId           string
	AssignedPartitions []topicPartitions
}

type consumerGroupDescription struct {
	Group    consumerGroup
	Protocol string
	State    string
	Members  []consumerGroupMember
}

type DescribeConsumerGroupFlags struct {
	ShowPartitionDetails bool
}

type GetConsumerGroupFlags struct {
	OutputFormat string
}

type ConsumerGroupOperation struct {
}

func (operation *ConsumerGroupOperation) DescribeConsumerGroup(flags DescribeConsumerGroupFlags, group string) {

	context := operations.CreateClientContext()

	var (
		err          error
		client       sarama.Client
		admin        sarama.ClusterAdmin
		descriptions []*sarama.GroupDescription
	)

	if client, err = operations.CreateClient(&context); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if admin, err = operations.CreateClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	if descriptions, err = admin.DescribeConsumerGroups([]string{group}); err != nil {
		output.Failf("failed to describe consumer group: %v", err)
	}

	for _, description := range descriptions {
		cg := consumerGroup{Name: description.GroupId, ProtocolType: description.ProtocolType}
		consumerGroupDescription := consumerGroupDescription{Group: cg, Protocol: description.Protocol, State: description.State, Members: make([]consumerGroupMember, 0)}

		for _, member := range description.Members {

			memberAssignment, err := member.GetMemberAssignment()

			if err != nil {
				output.Failf("failed to get group member assignment: %v", err)
			}

			topicPartitions := createTopicPartitions(memberAssignment.Topics)

			cgMember := consumerGroupMember{Host: member.ClientHost, ClientId: member.ClientId, AssignedPartitions: topicPartitions}

			consumerGroupDescription.Members = append(consumerGroupDescription.Members, cgMember)

			offsets, err := admin.ListConsumerGroupOffsets(description.GroupId, memberAssignment.Topics)

			if err != nil {
				output.Failf("failed to get group offsets: %v", err)
			}

			for i, topicPartition := range topicPartitions {

				topic := topicPartition.Topic

				for _, partition := range topicPartition.Partitions {

					offset, err := client.GetOffset(topic, partition, sarama.OffsetNewest)

					if err != nil {
						output.Failf("failed to get offset for topic %s partition %d: %v", topic, partition, err)
					}

					lag := offset - offsets.Blocks[topic][partition].Offset

					if flags.ShowPartitionDetails {
						detail := partitionDetail{Partition: partition, NewestOffset: offset, ConsumerOffset: offsets.Blocks[topic][partition].Offset, Lag: lag}
						topicPartitions[i].PartitionDetails = append(topicPartitions[i].PartitionDetails, detail)
						topicPartitions[i].Partitions = make([]int32, 0)
					}

					topicPartitions[i].TotalLag += lag
				}
			}

			output.PrintObject(consumerGroupDescription, "yaml")
		}
	}
}

func createTopicPartitions(topics map[string][]int32) []topicPartitions {

	topicPartitionList := make([]topicPartitions, 0)

	var topicsSorted []string
	for topic := range topics {
		topicsSorted = append(topicsSorted, topic)
	}
	sort.Strings(topicsSorted)

	for _, topic := range topicsSorted {
		if topic != "" {
			details := make([]partitionDetail, 0)
			topicPartitions := topicPartitions{Topic: topic, Partitions: topics[topic], PartitionDetails: details}
			topicPartitionList = append(topicPartitionList, topicPartitions)
		}
	}

	return topicPartitionList
}

func (operation *ConsumerGroupOperation) GetConsumerGroups(flags GetConsumerGroupFlags, topic string) {

	context := operations.CreateClientContext()

	var (
		err    error
		admin  sarama.ClusterAdmin
		groups map[string]string
	)

	if admin, err = operations.CreateClusterAdmin(&context); err != nil {
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

	if topic != "" {
		groupNames = filterGroups(admin, groupNames, topic)
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

		for _, member := range description.Members {

			metaData, err := member.GetMemberMetadata()

			if err != nil {
				output.Failf("failed to get group member metadata: %v", err)
			}

			if util.Contains(metaData.Topics, topic) {
				if !util.Contains(topicGroups, description.GroupId) {
					topicGroups = append(topicGroups, description.GroupId)
				}
			}
		}
	}

	return topicGroups
}
