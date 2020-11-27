package partitions

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
)

type AlterPartitionFlags struct {
	Replicas []int32
}

type PartitionOperation struct {
}

func (operation *PartitionOperation) AlterPartition(topic string, partition int32, flags AlterPartitionFlags) error {

	var (
		context operations.ClientContext
		client  sarama.Client
		admin   sarama.ClusterAdmin
		err     error
		exists  bool
	)

	if context, err = operations.CreateClientContext(); err != nil {
		return err
	}

	if client, err = operations.CreateClient(&context); err != nil {
		return err
	}

	if exists, err = operations.TopicExists(&client, topic); err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("topic '%s' does not exist", topic)
	}

	if admin, err = operations.CreateClusterAdmin(&context); err != nil {
		return err
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
				return errors.Errorf("unknown broker id to be used as replica: %d", replica)
			}
		}

		_ = admin
		//admin.AlterPartitionReassignments(topic, )
	}

	output.Infof(" hello replicas %v", flags.Replicas)
	return nil
}
