package operations

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
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

type DescribeTopicFlags struct {
	PrintConfigs bool
	OutputFormat string
}

type TopicOperation struct {
}

func (operation *TopicOperation) CreateTopics(topics []string, flags CreateTopicFlags) error {

	var (
		err     error
		context ClientContext
		admin   sarama.ClusterAdmin
	)

	if context, err = CreateClientContext(); err != nil {
		return err
	}

	if admin, err = CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
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
			return errors.Wrap(err, "failed to create topic")
		} else {
			output.Infof("topic created: %s", topic)
		}
	}
	return nil
}

func (operation *TopicOperation) DeleteTopics(topics []string) error {

	var (
		err     error
		context ClientContext
		admin   sarama.ClusterAdmin
	)

	if context, err = CreateClientContext(); err != nil {
		return err
	}

	if admin, err = CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	for _, topic := range topics {
		if err = admin.DeleteTopic(topic); err != nil {
			return errors.Wrap(err, "failed to delete topic")
		} else {
			output.Infof("topic deleted: %s", topic)
		}
	}
	return nil
}

func (operation *TopicOperation) DescribeTopic(topic string, flags DescribeTopicFlags) error {

	var (
		context ClientContext
		client  sarama.Client
		admin   sarama.ClusterAdmin
		err     error
		exists  bool
	)

	if context, err = CreateClientContext(); err != nil {
		return err
	}

	if client, err = CreateClient(&context); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if exists, err = TopicExists(&client, topic); err != nil {
		return errors.Wrap(err, "failed to read topics")
	}

	if !exists {
		return errors.Errorf("topic '%s' does not exist", topic)
	}

	if admin, err = CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	var t, _ = readTopic(&client, &admin, topic, allFields)

	if flags.PrintConfigs {
		if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
			t.Configs = nil
		} else {
			configTableWriter := output.CreateTableWriter()
			if err := configTableWriter.WriteHeader("CONFIG", "VALUE"); err != nil {
				return err
			}

			for _, c := range t.Configs {
				if err := configTableWriter.Write(c.Name, c.Value); err != nil {
					return err
				}
			}

			if err := configTableWriter.Flush(); err != nil {
				return err
			}
			output.PrintStrings("")
		}
	}

	partitionTableWriter := output.CreateTableWriter()

	if flags.OutputFormat == "" || flags.OutputFormat == "wide" {
		if err := partitionTableWriter.WriteHeader("PARTITION", "OLDEST_OFFSET", "NEWEST_OFFSET",
			"LEADER", "REPLICAS", "IN_SYNC_REPLICAS"); err != nil {
			return err
		}
	} else if flags.OutputFormat != "json" && flags.OutputFormat != "yaml" {
		return errors.Errorf("unknown outputFormat: %s", flags.OutputFormat)
	}

	if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
		return output.PrintObject(t, flags.OutputFormat)
	} else if flags.OutputFormat == "wide" || flags.OutputFormat == "" {
		for _, p := range t.Partitions {
			replicas := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(p.Replicas)), ","), "[]")
			inSyncReplicas := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(p.ISRs)), ","), "[]")
			if err := partitionTableWriter.Write(strconv.Itoa(int(p.Id)), strconv.Itoa(int(p.OldestOffset)),
				strconv.Itoa(int(p.NewestOffset)), p.Leader, replicas, inSyncReplicas); err != nil {
				return err
			}
		}
	}

	if flags.OutputFormat == "" || flags.OutputFormat == "wide" {
		if err := partitionTableWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (operation *TopicOperation) AlterTopic(topic string, flags AlterTopicFlags) error {

	var (
		context ClientContext
		client  sarama.Client
		admin   sarama.ClusterAdmin
		err     error
		exists  bool
	)

	if context, err = CreateClientContext(); err != nil {
		return err
	}

	if client, err = CreateClient(&context); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if exists, err = TopicExists(&client, topic); err != nil {
		return errors.Wrap(err, "failed to read topics")
	}

	if !exists {
		return errors.Errorf("topic '%s' does not exist", topic)
	}

	if admin, err = CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	var t, _ = readTopic(&client, &admin, topic, requestedTopicFields{partitionId: true, config: true})

	if flags.Partitions != 0 {
		if len(t.Partitions) > int(flags.Partitions) {
			return errors.New("Decreasing the number of partitions is not supported")
		}

		var emptyAssignment = make([][]int32, 0)

		err = admin.CreatePartitions(topic, flags.Partitions, emptyAssignment, flags.ValidateOnly)
		if err != nil {
			return errors.Errorf("Could not create partitions for topic '%s': %v", topic, err)
		}
	}

	if len(flags.Configs) == 0 {
		return operation.DescribeTopic(topic, DescribeTopicFlags{})
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
		return errors.Errorf("Could not alter topic config '%s': %v", topic, err)
	}

	return operation.DescribeTopic(topic, DescribeTopicFlags{})
}

func (operation *TopicOperation) GetTopics(flags GetTopicsFlags) error {

	var (
		err     error
		context ClientContext
		client  sarama.Client
		admin   sarama.ClusterAdmin
		topics  []string
	)

	if context, err = CreateClientContext(); err != nil {
		return err
	}

	if admin, err = CreateClusterAdmin(&context); err != nil {
		return errors.Wrap(err, "failed to create cluster admin")
	}

	if client, err = CreateClient(&context); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if topics, err = client.Topics(); err != nil {
		return errors.Wrap(err, "failed to read topics")
	}

	tableWriter := output.CreateTableWriter()
	var requestedFields requestedTopicFields

	if flags.OutputFormat == "" {
		requestedFields = requestedTopicFields{partitionId: true}
		if err := tableWriter.WriteHeader("TOPIC", "PARTITIONS"); err != nil {
			return err
		}
	} else if flags.OutputFormat == "compact" {
		tableWriter.Initialize()
	} else if flags.OutputFormat == "wide" {
		requestedFields = requestedTopicFields{partitionId: true, config: true}
		if err := tableWriter.WriteHeader("TOPIC", "PARTITIONS", "CONFIGS"); err != nil {
			return err
		}
	} else if flags.OutputFormat == "json" {
		requestedFields = allFields
	} else if flags.OutputFormat == "yaml" {
		requestedFields = allFields
	} else {
		return errors.Errorf("unknown outputFormat: %s", flags.OutputFormat)
	}

	topicChannel := make(chan topic)
	errChannel := make(chan error)

	// read topics in parallel
	for _, topic := range topics {
		go func(topic string) {
			t, err := readTopic(&client, &admin, topic, requestedFields)
			if err != nil {
				errChannel <- errors.Errorf("unable to read topic %s: %v", topic, err)
			} else {
				topicChannel <- t
			}
		}(topic)
	}

	topicList := make([]topic, 0, len(topics))
	for range topics {
		select {
		case topic := <-topicChannel:
			topicList = append(topicList, topic)
		case err := <-errChannel:
			return err
		}
	}

	sort.Slice(topicList, func(i, j int) bool {
		return topicList[i].Name < topicList[j].Name
	})

	if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
		return output.PrintObject(topicList, flags.OutputFormat)
	} else if flags.OutputFormat == "wide" {
		for _, t := range topicList {
			if err := tableWriter.Write(t.Name, strconv.Itoa(len(t.Partitions)), getConfigString(t.Configs)); err != nil {
				return err
			}
		}
	} else if flags.OutputFormat == "compact" {
		for _, t := range topicList {
			if err := tableWriter.Write(t.Name); err != nil {
				return err
			}
		}
	} else {
		for _, t := range topicList {
			if err := tableWriter.Write(t.Name, strconv.Itoa(len(t.Partitions))); err != nil {
				return err
			}
		}
	}

	if flags.OutputFormat == "wide" || flags.OutputFormat == "compact" || flags.OutputFormat == "" {
		if err := tableWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
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
	errChannel := make(chan error)

	// read partitions in parallel
	for _, p := range ps {

		go func(partitionId int32) {

			np := partition{Id: partitionId}

			if requestedFields.partitionOffset {
				if np.OldestOffset, err = (*client).GetOffset(name, partitionId, sarama.OffsetOldest); err != nil {
					errChannel <- errors.Errorf("unable to read oldest offset for topic %s partition %d", name, partitionId)
					return
				}

				if np.NewestOffset, err = (*client).GetOffset(name, partitionId, sarama.OffsetNewest); err != nil {
					errChannel <- errors.Errorf("unable to read newest offset for topic %s partition %d", name, partitionId)
					return
				}
			}

			if requestedFields.partitionLeader {
				if led, err = (*client).Leader(name, partitionId); err != nil {
					errChannel <- errors.Errorf("unable to read leader for topic %s partition %d", name, partitionId)
					return
				}
				np.Leader = led.Addr()
			}

			if requestedFields.partitionReplicas {
				if np.Replicas, err = (*client).Replicas(name, partitionId); err != nil {
					errChannel <- errors.Errorf("unable to read replicas for topic %s partition %d", name, partitionId)
					return
				}
				sort.Slice(np.Replicas, func(i, j int) bool { return np.Replicas[i] < np.Replicas[j] })
			}

			if requestedFields.partitionISRs {
				if np.ISRs, err = (*client).InSyncReplicas(name, partitionId); err != nil {
					errChannel <- errors.Errorf("unable to read inSyncReplicas for topic %s partition %d", name, partitionId)
					return
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
			return top, errors.Wrap(err, "failed to describe config")
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
