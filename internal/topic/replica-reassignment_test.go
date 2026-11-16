package topic

import (
	"errors"
	"slices"
	"testing"

	"github.com/IBM/sarama"
)

type reassignmentQueryAdmin struct {
	sarama.ClusterAdmin
	topic      string
	partitions []int32
	status     map[string]map[int32]*sarama.PartitionReplicaReassignmentsStatus
	err        error
}

func (a *reassignmentQueryAdmin) ListPartitionReassignments(topic string, partitions []int32) (map[string]map[int32]*sarama.PartitionReplicaReassignmentsStatus, error) {
	a.topic = topic
	a.partitions = slices.Clone(partitions)
	return a.status, a.err
}

func TestListTopicReassignmentsIncludesEveryPartition(t *testing.T) {
	for _, ids := range [][]int32{{0}, {0, 1, 2}, {2, 7, 11}, {}} {
		admin := &reassignmentQueryAdmin{}
		partitions := make([]Partition, len(ids))
		for i, id := range ids {
			partitions[i].ID = id
		}
		if _, err := listTopicReassignments(admin, "example-events", partitions); err != nil {
			t.Fatal(err)
		}
		if admin.topic != "example-events" || !slices.Equal(admin.partitions, ids) {
			t.Fatalf("status request: topic=%q partitions=%v, want partitions=%v", admin.topic, admin.partitions, ids)
		}
	}
}

func TestListTopicReassignmentsPreservesStatusAndErrors(t *testing.T) {
	pending := &sarama.PartitionReplicaReassignmentsStatus{Replicas: []int32{1, 2, 3}}
	admin := &reassignmentQueryAdmin{
		status: map[string]map[int32]*sarama.PartitionReplicaReassignmentsStatus{"example-events": {7: pending}},
	}
	status, err := listTopicReassignments(admin, "example-events", []Partition{{ID: 2}, {ID: 7}, {ID: 11}})
	if err != nil || status["example-events"][7] != pending {
		t.Fatalf("pending status was not preserved: status=%v err=%v", status, err)
	}
	cause := errors.New("synthetic status query failure")
	admin.err = cause
	if _, err := listTopicReassignments(admin, "example-events", []Partition{{ID: 7}}); !errors.Is(err, cause) {
		t.Fatalf("query error was not preserved: %v", err)
	}
}
