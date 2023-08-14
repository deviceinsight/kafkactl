package partition

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/internal"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type AlterPartitionFlags struct {
	Replicas     []int32
	ValidateOnly bool
}
type partition struct {
	ID           int32
	OldestOffset int64   `json:"oldestOffset" yaml:"oldestOffset"`
	NewestOffset int64   `json:"newestOffset" yaml:"newestOffset"`
	Leader       string  `json:",omitempty" yaml:",omitempty"`
	Replicas     []int32 `json:",omitempty" yaml:",omitempty,flow"`
	ISRs         []int32 `json:"inSyncReplicas,omitempty" yaml:"inSyncReplicas,omitempty,flow"`
}

type Operation struct {
}

func (operation *Operation) AlterPartition(topic string, partitionID int32, flags AlterPartitionFlags) error {

	var (
		context   internal.ClientContext
		client    sarama.Client
		admin     sarama.ClusterAdmin
		err       error
		exists    bool
		partition partition
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return err
	}

	if client, err = internal.CreateClient(&context); err != nil {
		return err
	}

	if exists, err = internal.TopicExists(&client, topic); err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("topic '%s' does not exist", topic)
	}

	if admin, err = internal.CreateClusterAdmin(&context); err != nil {
		return err
	}

	if partition, err = readPartition(&client, topic, partitionID); err != nil {
		return err
	}

	if len(flags.Replicas) > 0 {

		for _, replica := range flags.Replicas {
			brokerIDFound := false
			for _, broker := range client.Brokers() {
				if replica == broker.ID() {
					brokerIDFound = true
					break
				}
			}

			if !brokerIDFound {
				return errors.Errorf("unknown broker id to be used as replica: %d", replica)
			}
		}

		var replicaAssignment, err = readCurrentReplicas(&client, topic)
		if err != nil {
			return errors.Errorf("Unable to read current replicas for topic '%s': %v", topic, err)
		}

		replicaAssignment[partitionID] = flags.Replicas

		if flags.ValidateOnly {
			partition.Replicas = flags.Replicas
		} else {
			err = admin.AlterPartitionReassignments(topic, replicaAssignment)

			if err != nil {
				return errors.Errorf("Could not reassign partition replicas for topic '%s': %v", topic, err)
			}

			partitions := make([]int32, 1)
			partitions[0] = partitionID

			assignmentRunning := true

			for assignmentRunning {
				status, err := admin.ListPartitionReassignments(topic, partitions)
				if err != nil {
					return errors.Errorf("Could query reassignment status for topic '%s:%d': %v", topic, partitionID, err)
				}

				assignmentRunning = false

				if statusTopic, ok := status[topic]; ok {
					if statusPartition, ok := statusTopic[partitionID]; ok {
						output.Infof("reassignment running for topic=%s partition=%d: replicas:%v addingReplicas:%v removingReplicas:%v",
							topic, partitionID, statusPartition.Replicas, statusPartition.AddingReplicas, statusPartition.RemovingReplicas)
						time.Sleep(5 * time.Second)
						assignmentRunning = true
					}
				}
			}
			output.Infof("partition replicas have been reassigned")
		}
	}

	if flags.ValidateOnly {
		return printPartition(partition)
	}
	return nil
}

func printPartition(p partition) error {

	partitionTableWriter := output.CreateTableWriter()

	if err := partitionTableWriter.WriteHeader("PARTITION", "OLDEST_OFFSET", "NEWEST_OFFSET",
		"LEADER", "REPLICAS", "IN_SYNC_REPLICAS"); err != nil {
		return err
	}

	replicas := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(p.Replicas)), ","), "[]")
	inSyncReplicas := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(p.ISRs)), ","), "[]")
	if err := partitionTableWriter.Write(strconv.Itoa(int(p.ID)), strconv.Itoa(int(p.OldestOffset)),
		strconv.Itoa(int(p.NewestOffset)), p.Leader, replicas, inSyncReplicas); err != nil {
		return err
	}

	return partitionTableWriter.Flush()
}

func readPartition(client *sarama.Client, topic string, partitionID int32) (partition, error) {

	var (
		err error
		led *sarama.Broker
	)
	p := partition{ID: partitionID}

	if p.OldestOffset, err = (*client).GetOffset(topic, partitionID, sarama.OffsetOldest); err != nil {
		return p, errors.Errorf("unable to read oldest offset for topic %s partition %d", topic, partitionID)
	}

	if p.NewestOffset, err = (*client).GetOffset(topic, partitionID, sarama.OffsetNewest); err != nil {
		return p, errors.Errorf("unable to read newest offset for topic %s partition %d", topic, partitionID)
	}

	if led, err = (*client).Leader(topic, partitionID); err != nil {
		return p, errors.Errorf("unable to read leader for topic %s partition %d", topic, partitionID)
	}
	p.Leader = led.Addr()

	if p.Replicas, err = (*client).Replicas(topic, partitionID); err != nil {
		return p, errors.Errorf("unable to read replicas for topic %s partition %d", topic, partitionID)
	}
	sort.Slice(p.Replicas, func(i, j int) bool { return p.Replicas[i] < p.Replicas[j] })

	if p.ISRs, err = (*client).InSyncReplicas(topic, partitionID); err != nil {
		return p, errors.Errorf("unable to read inSyncReplicas for topic %s partition %d", topic, partitionID)
	}
	sort.Slice(p.ISRs, func(i, j int) bool { return p.ISRs[i] < p.ISRs[j] })

	return p, nil
}

func readCurrentReplicas(client *sarama.Client, topic string) ([][]int32, error) {
	var (
		err error
		ps  []int32
	)

	if ps, err = (*client).Partitions(topic); err != nil {
		return nil, err
	}

	partitionChannel := make(chan partition)
	errChannel := make(chan error)

	// read partitions in parallel
	for _, p := range ps {

		go func(partitionId int32) {

			np := partition{ID: partitionId}

			if np.Replicas, err = (*client).Replicas(topic, partitionId); err != nil {
				errChannel <- errors.Errorf("unable to read replicas for topic %s partition %d", topic, partitionId)
				return
			}
			sort.Slice(np.Replicas, func(i, j int) bool { return np.Replicas[i] < np.Replicas[j] })

			partitionChannel <- np
		}(p)
	}

	partitions := make([]partition, 0, len(ps))

	for range ps {
		select {
		case partition := <-partitionChannel:
			partitions = append(partitions, partition)
		case err := <-errChannel:
			return nil, err
		}
	}

	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].ID < partitions[j].ID
	})

	replicaAssignment := make([][]int32, 0, len(ps))

	for _, p := range partitions {
		replicaAssignment = append(replicaAssignment, p.Replicas)
	}

	return replicaAssignment, nil
}

func CompletePartitionIds(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {

	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var (
		context internal.ClientContext
		client  sarama.Client
		err     error
	)

	if context, err = internal.CreateClientContext(); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	if client, err = internal.CreateClient(&context); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	topicName := args[0]
	var partitions []int32

	if partitions, err = client.Partitions(topicName); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	partitionsString := make([]string, len(partitions))

	for i, p := range partitions {
		partitionsString[i] = strconv.Itoa(int(p))
	}

	return partitionsString, cobra.ShellCompDirectiveNoFileComp
}
