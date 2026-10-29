package topic

import "github.com/IBM/sarama"

func listTopicReassignments(admin sarama.ClusterAdmin, topic string, partitions []Partition) (map[string]map[int32]*sarama.PartitionReplicaReassignmentsStatus, error) {
	partitionIDs := make([]int32, len(partitions))
	for i, partition := range partitions {
		partitionIDs[i] = partition.ID
	}
	return admin.ListPartitionReassignments(topic, partitionIDs)
}
