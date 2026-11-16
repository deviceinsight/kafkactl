package topic

import (
	"testing"

	"github.com/IBM/sarama"
)

func TestGetTargetReplicasDecrease(t *testing.T) {

	// 3 brokers, current replicas on brokers 1,2,3, target RF=1
	// Should keep broker with lowest replica count
	brokerReplicaCount := map[int32]int{
		1: 2,
		2: 3,
		3: 1,
	}

	replicas, err := getTargetReplicas([]int32{1, 2, 3}, brokerReplicaCount, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(replicas) != 1 {
		t.Fatalf("expected 1 replica, got %d", len(replicas))
	}
	// Should keep broker 3 (lowest replica count)
	if replicas[0] != 3 {
		t.Fatalf("expected broker 3, got %d", replicas[0])
	}
}

func TestGetTargetReplicasIncrease(t *testing.T) {

	// 3 brokers, current replicas on broker 1 only, target RF=3
	brokerReplicaCount := map[int32]int{
		1: 2,
		2: 0,
		3: 1,
	}

	replicas, err := getTargetReplicas([]int32{1}, brokerReplicaCount, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(replicas) != 3 {
		t.Fatalf("expected 3 replicas, got %d", len(replicas))
	}
	// broker 1 was already there, brokers 2 and 3 should be added
	// broker 2 has the lowest count so should be added first
	if replicas[0] != 1 {
		t.Fatalf("expected broker 1 first, got %d", replicas[0])
	}
	if replicas[1] != 2 {
		t.Fatalf("expected broker 2 second, got %d", replicas[1])
	}
	if replicas[2] != 3 {
		t.Fatalf("expected broker 3 third, got %d", replicas[2])
	}
}

func TestGetTargetReplicasDecreaseWithUnavailableBroker(t *testing.T) {

	// Scenario from issue #256: broker 3 is unavailable
	// brokerReplicaCount only contains available brokers (1 and 2)
	// Current replicas include broker 3 (which is unavailable)
	brokerReplicaCount := map[int32]int{
		1: 2,
		2: 2,
	}

	// Partition has replicas on brokers 3 and 1
	// Since broker 3 is not in brokerReplicaCount (unavailable),
	// it should be removed first when decreasing
	replicas, err := getTargetReplicas([]int32{3, 1}, brokerReplicaCount, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(replicas) != 1 {
		t.Fatalf("expected 1 replica, got %d", len(replicas))
	}
	// Should keep broker 1 (available) and remove broker 3 (unavailable)
	if replicas[0] != 1 {
		t.Fatalf("expected broker 1, got %d", replicas[0])
	}
}

func TestGetTargetReplicasDecreaseMultipleUnavailableBrokers(t *testing.T) {

	// Multiple unavailable brokers should all be removed first
	brokerReplicaCount := map[int32]int{
		1: 3,
		2: 3,
	}

	// Replicas on brokers 3, 4, 1, 2 - brokers 3 and 4 are unavailable
	// Target RF=2, should remove unavailable brokers first
	replicas, err := getTargetReplicas([]int32{3, 4, 1, 2}, brokerReplicaCount, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(replicas) != 2 {
		t.Fatalf("expected 2 replicas, got %d", len(replicas))
	}
	// Should keep brokers 1 and 2 (available) and remove 3 and 4 (unavailable)
	for _, r := range replicas {
		if r == 3 || r == 4 {
			t.Fatalf("unavailable broker %d should have been removed", r)
		}
	}
}

func TestGetTargetReplicasDecreaseAllUnavailable(t *testing.T) {

	// Edge case: all current replicas are on unavailable brokers
	brokerReplicaCount := map[int32]int{
		1: 0,
		2: 0,
	}

	// Current replicas are on brokers 3 and 4 (both unavailable)
	replicas, err := getTargetReplicas([]int32{3, 4}, brokerReplicaCount, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still work, keeping one of them
	if len(replicas) != 1 {
		t.Fatalf("expected 1 replica, got %d", len(replicas))
	}
}

func TestGetTargetReplicasDecreaseUnavailableThenIncrease(t *testing.T) {

	// All current replicas on unavailable brokers; result must use only available brokers.
	brokerReplicaCount := map[int32]int{
		1: 0,
		2: 0,
	}

	replicas, err := getTargetReplicas([]int32{3, 4, 5}, brokerReplicaCount, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 2 {
		t.Fatalf("expected 2 replicas, got %d", len(replicas))
	}
	for _, r := range replicas {
		if r == 3 || r == 4 || r == 5 {
			t.Fatalf("unavailable broker %d should not be in result %v", r, replicas)
		}
	}
}

func TestGetTargetReplicasNotEnoughBrokers(t *testing.T) {

	brokerReplicaCount := map[int32]int{
		1: 1,
	}

	_, err := getTargetReplicas([]int32{1}, brokerReplicaCount, 3)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if err.Error() != "not enough brokers" {
		t.Fatalf("expected 'not enough brokers' error, got: %v", err)
	}
}

// sortReadClient is a minimal stub that returns replica IDs together with
// ErrReplicaNotAvailable, as sarama does for a partition whose replica broker
// is offline.
type sortReadClient struct {
	sarama.Client
}

func (c *sortReadClient) Partitions(string) ([]int32, error) {
	return []int32{0}, nil
}

func (c *sortReadClient) Replicas(string, int32) ([]int32, error) {
	return []int32{3, 1, 2}, sarama.ErrReplicaNotAvailable
}

func TestReadTopicSortsReplicasEvenWhenBrokerUnavailable(t *testing.T) {
	var client sarama.Client = &sortReadClient{}
	fields := requestedTopicFields{partitionID: true, partitionReplicas: true, config: NoConfigs}
	topic, err := readTopic(&client, nil, "example-events", fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topic.ReplicationFactor != 3 {
		t.Fatalf("expected RF=3, got %d", topic.ReplicationFactor)
	}
	replicas := topic.Partitions[0].Replicas
	for i := 1; i < len(replicas); i++ {
		if replicas[i-1] >= replicas[i] {
			t.Fatalf("replicas not sorted even though a replica broker is unavailable: %v", replicas)
		}
	}
}
