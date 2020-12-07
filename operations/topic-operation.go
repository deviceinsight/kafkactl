package operations

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sort"
	"strconv"
	"strings"
	"time"
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
	Partitions        int32
	ReplicationFactor int16
	ValidateOnly      bool
	Configs           []string
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

	return operation.printTopic(t, flags)
}

func (operation *TopicOperation) printTopic(topic topic, flags DescribeTopicFlags) error {

	if flags.PrintConfigs {
		if flags.OutputFormat == "json" || flags.OutputFormat == "yaml" {
			topic.Configs = nil
		} else {
			configTableWriter := output.CreateTableWriter()
			if err := configTableWriter.WriteHeader("CONFIG", "VALUE"); err != nil {
				return err
			}

			for _, c := range topic.Configs {
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
		return output.PrintObject(topic, flags.OutputFormat)
	} else if flags.OutputFormat == "wide" || flags.OutputFormat == "" {
		for _, p := range topic.Partitions {
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

	var t, _ = readTopic(&client, &admin, topic, allFields)

	if flags.Partitions != 0 {
		if len(t.Partitions) > int(flags.Partitions) {
			return errors.New("Decreasing the number of partitions is not supported")
		}

		if flags.ValidateOnly {
			for len(t.Partitions) < int(flags.Partitions) {
				t.Partitions = append(t.Partitions, partition{Id: int32(len(t.Partitions)), NewestOffset: 0, OldestOffset: 0})
			}
		} else {
			var emptyAssignment = make([][]int32, 0)

			err = admin.CreatePartitions(topic, flags.Partitions, emptyAssignment, flags.ValidateOnly)
			if err != nil {
				return errors.Errorf("Could not create partitions for topic '%s': %v", topic, err)
			}
		}
	}

	if flags.ReplicationFactor > 0 {

		var brokers = client.Brokers()

		if int(flags.ReplicationFactor) > len(brokers) {
			return errors.Errorf("Replication factor for topic '%s' must not exceed the number of available brokers", topic)
		}

		brokerReplicaCount := make(map[int32]int)
		for _, broker := range brokers {
			brokerReplicaCount[broker.ID()] = 0
		}

		for _, partition := range t.Partitions {
			for _, brokerId := range partition.Replicas {
				brokerReplicaCount[brokerId] += 1
			}
		}

		var replicaAssignment = make([][]int32, 0, int16(len(t.Partitions)))

		for _, partition := range t.Partitions {

			var replicas, err = getTargetReplicas(partition.Replicas, brokerReplicaCount, flags.ReplicationFactor)
			if err != nil {
				return errors.Wrap(err, "unable to determine target replicas")
			}
			replicaAssignment = append(replicaAssignment, replicas)
		}

		for brokerId, replicaCount := range brokerReplicaCount {
			output.Debugf("broker %d now has %d replicas", brokerId, replicaCount)
		}

		if flags.ValidateOnly {
			for i := range t.Partitions {
				t.Partitions[i].Replicas = replicaAssignment[i]
			}
		} else {
			err = admin.AlterPartitionReassignments(topic, replicaAssignment)
			if err != nil {
				return errors.Errorf("Could not reassign partition replicas for topic '%s': %v", topic, err)
			}

			partitions := make([]int32, len(t.Partitions))

			for _, p := range t.Partitions {
				partitions[0] = p.Id
			}

			assignmentRunning := true

			for assignmentRunning {
				status, err := admin.ListPartitionReassignments(topic, partitions)
				if err != nil {
					return errors.Errorf("Could query reassignment status for topic '%s': %v", topic, err)
				}

				assignmentRunning = false

				if statusTopic, ok := status[topic]; ok {
					for partitionId, statusPartition := range statusTopic {
						output.Debugf("Reassignment running: %s:%d replicas: %v addingReplicas: %v removingReplicas: %v",
							topic, partitionId, statusPartition.Replicas, statusPartition.AddingReplicas, statusPartition.RemovingReplicas)
						time.Sleep(2 * time.Second)
						assignmentRunning = true
					}
				} else {
					output.Debugf("Emtpy list partition reassignment result returned (len status: %d)", len(status))
				}
			}
		}
	}

	if len(flags.Configs) > 0 {
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

		if flags.ValidateOnly {
			// validate only - directly alter the response object
			t.Configs = make([]config, 0, len(mergedConfigEntries))
			for key, value := range mergedConfigEntries {
				t.Configs = append(t.Configs, config{Name: key, Value: *value})
			}
		} else {
			if err = admin.AlterConfig(sarama.TopicResource, topic, mergedConfigEntries, flags.ValidateOnly); err != nil {
				return errors.Errorf("Could not alter topic config '%s': %v", topic, err)
			}
		}
	}

	if !flags.ValidateOnly {
		t, _ = readTopic(&client, &admin, topic, allFields)
	}

	describeFlags := DescribeTopicFlags{PrintConfigs: len(flags.Configs) > 0}
	return operation.printTopic(t, describeFlags)
}

func (operation *TopicOperation) ListTopicsNames() ([]string, error) {

	var (
		err     error
		context ClientContext
		client  sarama.Client
		topics  []string
	)

	if context, err = CreateClientContext(); err != nil {
		return nil, err
	}

	if client, err = CreateClient(&context); err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	if topics, err = client.Topics(); err != nil {
		return nil, errors.Wrap(err, "failed to read topics")
	} else {
		return topics, nil
	}
}

func getTargetReplicas(currentReplicas []int32, brokerReplicaCount map[int32]int, targetReplicationFactor int16) ([]int32, error) {

	replicas := currentReplicas

	for len(replicas) > int(targetReplicationFactor) {

		sort.Slice(replicas, func(i, j int) bool {
			brokerI := replicas[i]
			brokerJ := replicas[j]
			return brokerReplicaCount[brokerI] < brokerReplicaCount[brokerJ] || (brokerReplicaCount[brokerI] == brokerReplicaCount[brokerJ] && brokerI < brokerJ)
		})

		lastReplica := replicas[len(replicas)-1]
		replicas = replicas[:len(replicas)-1]
		brokerReplicaCount[lastReplica] -= 1
	}

	var unusedBrokerIds []int32

	if len(replicas) < int(targetReplicationFactor) {
		for brokerId := range brokerReplicaCount {
			if !util.ContainsInt32(replicas, brokerId) {
				unusedBrokerIds = append(unusedBrokerIds, brokerId)
			}
		}
		if len(unusedBrokerIds) < (int(targetReplicationFactor) - len(replicas)) {
			return nil, errors.New("not enough brokers")
		}
	}

	for len(replicas) < int(targetReplicationFactor) {

		sort.Slice(unusedBrokerIds, func(i, j int) bool {
			brokerI := unusedBrokerIds[i]
			brokerJ := unusedBrokerIds[j]
			return brokerReplicaCount[brokerI] < brokerReplicaCount[brokerJ] || (brokerReplicaCount[brokerI] == brokerReplicaCount[brokerJ] && brokerI > brokerJ)
		})

		replicas = append(replicas, unusedBrokerIds[0])
		brokerReplicaCount[unusedBrokerIds[0]] += 1
		unusedBrokerIds = unusedBrokerIds[1:]
	}

	return replicas, nil
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
				} else {
					np.Leader = led.Addr()
				}
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
		select {
		case partition := <-partitionChannel:
			top.Partitions = append(top.Partitions, partition)
		case err := <-errChannel:
			return top, err
		}
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

func CompleteTopicNames(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	topics, err := (&TopicOperation{}).ListTopicsNames()

	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return topics, cobra.ShellCompDirectiveNoFileComp
}
