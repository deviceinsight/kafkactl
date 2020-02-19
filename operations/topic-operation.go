package operations

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"sort"
	"strconv"
	"strings"
)

type topic struct {
	Name       string
	Partitions []partition `json:",omitempty" yaml:",omitempty"`
	Configs    []config    `json:",omitempty" yaml:",omitempty"`
}

type partition struct {
	Id           int32
	OldestOffset int64   `json:"oldestOffset" yaml:"oldestOffset"`
	NewestOffset int64   `json:"newestOffset" yaml:"newestOffset"`
	Leader       string  `json:",omitempty" yaml:",omitempty"`
	Replicas     []int32 `json:",omitempty" yaml:",omitempty,flow"`
	ISRs         []int32 `json:"inSyncReplicas,omitempty" yaml:"inSyncReplicas,omitempty,flow"`
}

type requestedTopicFields struct {
	partitionId       bool
	partitionOffset   bool
	partitionLeader   bool
	partitionReplicas bool
	partitionISRs     bool
	config            bool
}

var allFields = requestedTopicFields{partitionId: true, partitionOffset: true, partitionLeader: true, partitionReplicas: true, partitionISRs: true, config: true}

type config struct {
	Name  string
	Value string
}

type GetTopicsFlags struct {
	OutputFormat string
}

type CreateTopicFlags struct {
	Partitions        int32
	ReplicationFactor int16
	ValidateOnly      bool
	Configs           []string
}

type AlterTopicFlags struct {
	Partitions   int32
	ValidateOnly bool
	Configs      []string
}

type TopicOperation struct {
}

func (operation *TopicOperation) CreateTopics(topics []string, flags CreateTopicFlags) {

	context := CreateClientContext()

	var (
		err   error
		admin sarama.ClusterAdmin
	)

	if admin, err = CreateClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	topicDetails := sarama.TopicDetail{
		NumPartitions:     flags.Partitions,
		ReplicationFactor: flags.ReplicationFactor,
		ConfigEntries:     map[string]*string{},
	}

	for _, config := range flags.Configs {
		configParts := strings.Split(config, "=")
		topicDetails.ConfigEntries[configParts[0]] = &configParts[1]
	}

	for _, topic := range topics {
		if err = admin.CreateTopic(topic, &topicDetails, flags.ValidateOnly); err != nil {
			output.Failf("failed to create topic: %v", err)
		} else {
			output.Infof("topic created: %s", topic)
		}
	}
}

func (operation *TopicOperation) DeleteTopics(topics []string) {

	context := CreateClientContext()

	var (
		err   error
		admin sarama.ClusterAdmin
	)

	if admin, err = CreateClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	for _, topic := range topics {
		if err = admin.DeleteTopic(topic); err != nil {
			output.Failf("failed to delete topic: %v", err)
		} else {
			output.Infof("topic deleted: %s", topic)
		}
	}
}

func (operation *TopicOperation) DescribeTopic(topic string) {

	context := CreateClientContext()

	var (
		client sarama.Client
		admin  sarama.ClusterAdmin
		err    error
		exists bool
	)

	if client, err = CreateClient(&context); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if exists, err = TopicExists(&client, topic); err != nil {
		output.Failf("failed to read topics err=%v", err)
	}

	if !exists {
		output.Failf("topic '%s' does not exist", topic)
	}

	if admin, err = CreateClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	var t, _ = readTopic(&client, &admin, topic, allFields)
	output.PrintObject(t, "yaml")
}

func (operation *TopicOperation) AlterTopic(topic string, flags AlterTopicFlags) {

	context := CreateClientContext()

	var (
		client sarama.Client
		admin  sarama.ClusterAdmin
		err    error
		exists bool
	)

	if client, err = CreateClient(&context); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if exists, err = TopicExists(&client, topic); err != nil {
		output.Failf("failed to read topics err=%v", err)
	}

	if !exists {
		output.Failf("topic '%s' does not exist", topic)
	}

	if admin, err = CreateClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	var t, _ = readTopic(&client, &admin, topic, requestedTopicFields{partitionId: true, config: true})

	if flags.Partitions != 0 {
		if len(t.Partitions) > int(flags.Partitions) {
			output.Failf("Decreasing the number of partitions is not supported")
		}

		var emptyAssignment = make([][]int32, 0, 0)

		err = admin.CreatePartitions(topic, flags.Partitions, emptyAssignment, flags.ValidateOnly)
		if err != nil {
			output.Failf("Could not create partitions for topic '%s': %v", topic, err)
		}
	}

	if len(flags.Configs) == 0 {
		operation.DescribeTopic(topic)
		return
	}

	mergedConfigEntries := make(map[string]*string)

	for i, config := range t.Configs {
		mergedConfigEntries[config.Name] = &(t.Configs[i].Value)
	}

	for _, config := range flags.Configs {
		configParts := strings.Split(config, "=")

		if len(configParts) == 2 {
			if len(configParts[1]) == 0 {
				delete(mergedConfigEntries, configParts[0])
			} else {
				mergedConfigEntries[configParts[0]] = &configParts[1]
			}
		}
	}

	if err = admin.AlterConfig(sarama.TopicResource, topic, mergedConfigEntries, flags.ValidateOnly); err != nil {
		output.Failf("Could not alter topic config '%s': %v", topic, err)
	}

	operation.DescribeTopic(topic)
}

func (operation *TopicOperation) GetTopics(flags GetTopicsFlags) {

	context := CreateClientContext()

	var (
		err    error
		client sarama.Client
		admin  sarama.ClusterAdmin
		topics []string
	)

	tableWriter := output.CreateTableWriter()

	if admin, err = CreateClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	if client, err = CreateClient(&context); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if topics, err = client.Topics(); err != nil {
		output.Failf("failed to read topics err=%v", err)
	}

	var requestedFields requestedTopicFields

	if flags.OutputFormat == "" {
		requestedFields = requestedTopicFields{partitionId: true}
		tableWriter.WriteHeader("TOPIC", "PARTITIONS")
	} else if flags.OutputFormat == "compact" {
		tableWriter.Initialize()
	} else if flags.OutputFormat == "wide" {
		requestedFields = requestedTopicFields{partitionId: true, config: true}
		tableWriter.WriteHeader("TOPIC", "PARTITIONS", "CONFIGS")
	} else if flags.OutputFormat == "json" {
		requestedFields = allFields
	} else if flags.OutputFormat == "yaml" {
		requestedFields = allFields
	} else {
		output.Failf("unknown outputFormat: %s", flags.OutputFormat)
	}

	topicChannel := make(chan topic)

	// read topics in parallel
	for _, topic := range topics {
		go func(topic string) {
			t, err := readTopic(&client, &admin, topic, requestedFields)
			if err != nil {
				output.Failf("unable to read topic %s: %v", topic, err)
			}
			topicChannel <- t
		}(topic)
	}

	topicList := make([]topic, 0, len(topics))
	for range topics {
		topicList = append(topicList, <-topicChannel)
	}

	sort.Slice(topicList, func(i, j int) bool {
		return topicList[i].Name < topicList[j].Name
	})

	if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
		output.PrintObject(topicList, flags.OutputFormat)
	} else if flags.OutputFormat == "wide" {
		for _, t := range topicList {
			tableWriter.Write(t.Name, strconv.Itoa(len(t.Partitions)), getConfigString(t.Configs))
		}
	} else if flags.OutputFormat == "compact" {
		for _, t := range topicList {
			tableWriter.Write(t.Name)
		}
	} else {
		for _, t := range topicList {
			tableWriter.Write(t.Name, strconv.Itoa(len(t.Partitions)))
		}
	}

	if flags.OutputFormat == "wide" || flags.OutputFormat == "compact" || flags.OutputFormat == "" {
		tableWriter.Flush()
	}
}

func readTopic(client *sarama.Client, admin *sarama.ClusterAdmin, name string, requestedFields requestedTopicFields) (topic, error) {
	var (
		err           error
		ps            []int32
		led           *sarama.Broker
		configEntries []sarama.ConfigEntry
		top           = topic{Name: name}
	)

	if !requestedFields.partitionId {
		return top, nil
	}

	if ps, err = (*client).Partitions(name); err != nil {
		return top, err
	}

	partitionChannel := make(chan partition)

	// read partitions in parallel
	for _, p := range ps {

		go func(partitionId int32) {

			np := partition{Id: partitionId}

			if requestedFields.partitionOffset {
				if np.OldestOffset, err = (*client).GetOffset(name, partitionId, sarama.OffsetOldest); err != nil {
					output.Failf("unable to read oldest offset for topic %s partition %d", name, partitionId)
				}

				if np.NewestOffset, err = (*client).GetOffset(name, partitionId, sarama.OffsetNewest); err != nil {
					output.Failf("unable to read newest offset for topic %s partition %d", name, partitionId)
				}
			}

			if requestedFields.partitionLeader {
				if led, err = (*client).Leader(name, partitionId); err != nil {
					output.Failf("unable to read leader for topic %s partition %d", name, partitionId)
				}
				np.Leader = led.Addr()
			}

			if requestedFields.partitionReplicas {
				if np.Replicas, err = (*client).Replicas(name, partitionId); err != nil {
					output.Failf("unable to read replicas for topic %s partition %d", name, partitionId)
				}
				sort.Slice(np.Replicas, func(i, j int) bool { return np.Replicas[i] < np.Replicas[j] })
			}

			if requestedFields.partitionISRs {
				if np.ISRs, err = (*client).InSyncReplicas(name, partitionId); err != nil {
					output.Failf("unable to read inSyncReplicas for topic %s partition %d", name, partitionId)
				}
				sort.Slice(np.ISRs, func(i, j int) bool { return np.ISRs[i] < np.ISRs[j] })
			}

			partitionChannel <- np
		}(p)
	}

	for range ps {
		top.Partitions = append(top.Partitions, <-partitionChannel)
	}

	sort.Slice(top.Partitions, func(i, j int) bool {
		return top.Partitions[i].Id < top.Partitions[j].Id
	})

	if requestedFields.config {

		configResource := sarama.ConfigResource{
			Type: sarama.TopicResource,
			Name: name,
		}

		if configEntries, err = (*admin).DescribeConfig(configResource); err != nil {
			output.Failf("failed to describe config: %v", err)
		}

		for _, configEntry := range configEntries {

			if !configEntry.Default && configEntry.Source != sarama.SourceDefault {
				entry := config{Name: configEntry.Name, Value: configEntry.Value}
				top.Configs = append(top.Configs, entry)
			}
		}
	}

	return top, nil
}

func getConfigString(configs []config) string {

	var configStrings []string

	for _, config := range configs {
		configStrings = append(configStrings, fmt.Sprintf("%s=%s", config.Name, config.Value))
	}

	return strings.Trim(strings.Join(configStrings, ","), "[]")
}
