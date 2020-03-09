package partitions

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
)

type AlterPartitionFlags struct {
	Replicas []int32
}

type PartitionOperation struct {
}

func (operation *PartitionOperation) AlterPartition(topic string, partition int32, flags AlterPartitionFlags) {

	context := operations.CreateClientContext()

	var (
		client sarama.Client
		admin  sarama.ClusterAdmin
		err    error
		exists bool
	)

	if client, err = operations.CreateClient(&context); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if exists, err = operations.TopicExists(&client, topic); err != nil {
		output.Failf("failed to read topics err=%v", err)
	}

	if !exists {
		output.Failf("topic '%s' does not exist", topic)
	}

	if admin, err = operations.CreateClusterAdmin(&context); err != nil {
		output.Failf("failed to create cluster admin: %v", err)
	}

	if len(flags.Replicas) > 0 {

		for _, replica := range flags.Replicas {
			brokerIdFound := false
			for _, broker := range client.Brokers() {
				if replica == broker.ID() {
					brokerIdFound = true
					break
				}
			}

			if !brokerIdFound {
				output.Failf("unknown broker id to be used as replica: %d", replica)
			}
		}

		_ = admin
		//admin.AlterPartitionReassignments(topic, )
	}

	output.Infof(" hello replicas %v", flags.Replicas)
}
