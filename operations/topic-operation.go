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
	OldestOffset int64
	NewestOffset int64
	Leader       string  `json:",omitempty" yaml:",omitempty"`
	Replicas     []int32 `json:",omitempty" yaml:",omitempty"`
	ISRs         []int32 `json:",omitempty" yaml:",omitempty"`
}

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

	if admin, err = createClusterAdmin(&context); err != nil {
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

	if admin, err = createClusterAdmin(&context); err != nil {
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

	if admin, err = createClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	var t, _ = readTopic(&client, &admin, topic, true, false, true, true)
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

	if admin, err = createClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	var t, _ = readTopic(&client, &admin, topic, true, false, false, true)

	if flags.Partitions != 0 {
		if len(t.Partitions) > int(flags.Partitions) {
			output.Failf("Decreasing the number of partitions is not supported")
		}

		var emptyAssignment [][]int32 = make([][]int32, 0, 0)

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

	if admin, err = createClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	if client, err = CreateClient(&context); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if topics, err = client.Topics(); err != nil {
		output.Failf("failed to read topics err=%v", err)
	}

	sort.Strings(topics)

	if flags.OutputFormat == "" {
		tableWriter.WriteHeader("TOPIC", "PARTITIONS")
	} else if flags.OutputFormat == "compact" {
		output.PrintStrings(topics...)
		return
	} else if flags.OutputFormat == "wide" {
		tableWriter.WriteHeader("TOPIC", "PARTITIONS", "CONFIGS")
	}

	for _, topic := range topics {
		var t, _ = readTopic(&client, &admin, topic, true, false, true, true)
		if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
			output.PrintObject(t, flags.OutputFormat)
		} else if flags.OutputFormat == "wide" {
			tableWriter.Write(t.Name, strconv.Itoa(len(t.Partitions)), getConfigString(t.Configs))
		} else {
			tableWriter.Write(t.Name, strconv.Itoa(len(t.Partitions)))
		}
	}

	if flags.OutputFormat == "wide" || flags.OutputFormat == "" {
		tableWriter.Flush()
	}
}

func readTopic(client *sarama.Client, admin *sarama.ClusterAdmin, name string, readPartitions bool, readLeaders bool, readReplicas bool, readConfigs bool) (topic, error) {
	var (
		err           error
		ps            []int32
		led           *sarama.Broker
		configEntries []sarama.ConfigEntry
		top           = topic{Name: name}
	)

	if !readPartitions {
		return top, nil
	}

	if ps, err = (*client).Partitions(name); err != nil {
		return top, err
	}

	for _, p := range ps {
		np := partition{Id: p}

		if np.OldestOffset, err = (*client).GetOffset(name, p, sarama.OffsetOldest); err != nil {
			return top, err
		}

		if np.NewestOffset, err = (*client).GetOffset(name, p, sarama.OffsetNewest); err != nil {
			return top, err
		}

		if readLeaders {
			if led, err = (*client).Leader(name, p); err != nil {
				return top, err
			}
			np.Leader = led.Addr()
		}

		if readReplicas {
			if np.Replicas, err = (*client).Replicas(name, p); err != nil {
				return top, err
			}

			if np.ISRs, err = (*client).InSyncReplicas(name, p); err != nil {
				return top, err
			}
		}

		top.Partitions = append(top.Partitions, np)
	}

	if readConfigs {

		configResource := sarama.ConfigResource{
			Type: sarama.TopicResource,
			Name: name,
		}

		if configEntries, err = (*admin).DescribeConfig(configResource); err != nil {
			output.Failf("failed to describe config: %v", err)
		}

		for _, configEntry := range configEntries {

			if !configEntry.Default {
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
