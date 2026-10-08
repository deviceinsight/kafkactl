package topic

import (
	"testing"
)

func TestGetTargetReplicasDecrease(t *testing.T) {

	// 3 brokers, current replicas on brokers 1,2,3, target RF=1
	// Should keep broker with lowest replica count
	brokerReplicaCount := map[int32]int{
		1: 2,
		2: 3,
		3: 1,
	}

	replicas, err := getTargetReplicas([]int32{1, 2, 3}, brokerReplicaCount, 1, nil)
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

	replicas, err := getTargetReplicas([]int32{1}, brokerReplicaCount, 3, nil)
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
	replicas, err := getTargetReplicas([]int32{3, 1}, brokerReplicaCount, 1, nil)
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
	replicas, err := getTargetReplicas([]int32{3, 4, 1, 2}, brokerReplicaCount, 2, nil)
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
	replicas, err := getTargetReplicas([]int32{3, 4}, brokerReplicaCount, 1, nil)
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

	replicas, err := getTargetReplicas([]int32{3, 4, 5}, brokerReplicaCount, 2, nil)
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

	_, err := getTargetReplicas([]int32{1}, brokerReplicaCount, 3, nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if err.Error() != "not enough brokers" {
		t.Fatalf("expected 'not enough brokers' error, got: %v", err)
	}
}

func TestGetTargetReplicasIncreasePrefersUnusedRack(t *testing.T) {

	// Unused rack wins over a broker with a lower replica count on a used rack.
	brokerReplicaCount := map[int32]int{
		1: 5,
		2: 0,
		3: 4,
	}
	brokerRacks := map[int32]string{
		1: "a",
		2: "a",
		3: "b",
	}

	replicas, err := getTargetReplicas([]int32{1}, brokerReplicaCount, 2, brokerRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 2 || replicas[0] != 1 || replicas[1] != 3 {
		t.Fatalf("expected [1 3], got %v", replicas)
	}
}

func TestGetTargetReplicasIncreaseBalanceOnUsedRack(t *testing.T) {

	// Every candidate shares the partition's rack, so the lower replica count wins.
	brokerReplicaCount := map[int32]int{
		1: 1,
		2: 0,
		3: 2,
	}
	brokerRacks := map[int32]string{
		1: "a",
		2: "a",
		3: "a",
	}

	replicas, err := getTargetReplicas([]int32{1}, brokerReplicaCount, 2, brokerRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 2 || replicas[0] != 1 || replicas[1] != 2 {
		t.Fatalf("expected [1 2], got %v", replicas)
	}
}

func TestGetTargetReplicasIncreaseFillsDistinctRacks(t *testing.T) {

	// Distinct unused racks are filled before a rack is used twice, even when the
	// second unused rack has a higher replica count.
	brokerReplicaCount := map[int32]int{
		1: 0,
		2: 0,
		3: 0,
		4: 5,
	}
	brokerRacks := map[int32]string{
		1: "a",
		2: "a",
		3: "b",
		4: "c",
	}

	replicas, err := getTargetReplicas([]int32{1}, brokerReplicaCount, 3, brokerRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 3 || replicas[0] != 1 || replicas[1] != 3 || replicas[2] != 4 {
		t.Fatalf("expected [1 3 4], got %v", replicas)
	}
}

func TestGetTargetReplicasIncreaseKeepsExistingOrder(t *testing.T) {

	brokerReplicaCount := map[int32]int{
		1: 0,
		2: 1,
		3: 0,
	}
	brokerRacks := map[int32]string{
		1: "b",
		2: "a",
		3: "c",
	}

	replicas, err := getTargetReplicas([]int32{2, 1}, brokerReplicaCount, 3, brokerRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 3 || replicas[0] != 2 || replicas[1] != 1 || replicas[2] != 3 {
		t.Fatalf("expected [2 1 3], got %v", replicas)
	}
}

func TestGetTargetReplicasRackAwareAcrossPartitions(t *testing.T) {

	// Partitions share one count map. The higher broker ID wins a count tie, and
	// the updated count is visible to the next partition.
	brokerReplicaCount := map[int32]int{
		1: 1,
		2: 1,
		3: 0,
		4: 0,
	}
	brokerRacks := map[int32]string{
		1: "a",
		2: "a",
		3: "b",
		4: "b",
	}

	first, err := getTargetReplicas([]int32{1}, brokerReplicaCount, 2, brokerRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(first) != 2 || first[0] != 1 || first[1] != 4 {
		t.Fatalf("expected [1 4], got %v", first)
	}

	second, err := getTargetReplicas([]int32{2}, brokerReplicaCount, 2, brokerRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(second) != 2 || second[0] != 2 || second[1] != 3 {
		t.Fatalf("expected [2 3], got %v", second)
	}
}

func TestGetTargetReplicasEmptyRackStringsMatchBalanceOnly(t *testing.T) {

	decreaseCounts := map[int32]int{
		1: 2,
		2: 3,
		3: 1,
	}
	emptyRacks := map[int32]string{
		1: "",
		2: "",
		3: "",
	}

	replicas, err := getTargetReplicas([]int32{1, 2, 3}, decreaseCounts, 1, emptyRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 1 || replicas[0] != 3 {
		t.Fatalf("expected [3], got %v", replicas)
	}

	increaseCounts := map[int32]int{
		1: 2,
		2: 0,
		3: 1,
	}
	replicas, err = getTargetReplicas([]int32{1}, increaseCounts, 3, emptyRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 3 || replicas[0] != 1 || replicas[1] != 2 || replicas[2] != 3 {
		t.Fatalf("expected [1 2 3], got %v", replicas)
	}
}

func TestGetTargetReplicasDecreaseDropsDuplicatedRack(t *testing.T) {

	// A duplicated rack is removed before a unique rack, ahead of replica count.
	brokerReplicaCount := map[int32]int{
		1: 10,
		2: 0,
		3: 0,
	}
	brokerRacks := map[int32]string{
		1: "a",
		2: "a",
		3: "b",
	}

	replicas, err := getTargetReplicas([]int32{1, 2, 3}, brokerReplicaCount, 2, brokerRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 2 || replicas[0] != 3 || replicas[1] != 2 {
		t.Fatalf("expected [3 2], got %v", replicas)
	}
}

func TestGetTargetReplicasDecreaseUnavailableBeforeUniqueRack(t *testing.T) {

	// An unavailable broker is removed first even when it is the only member of its rack.
	brokerReplicaCount := map[int32]int{
		1: 1,
	}
	brokerRacks := map[int32]string{
		1: "a",
		3: "b",
	}

	replicas, err := getTargetReplicas([]int32{3, 1}, brokerReplicaCount, 1, brokerRacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 1 || replicas[0] != 1 {
		t.Fatalf("expected [1], got %v", replicas)
	}
}

func TestGetTargetReplicasDecreaseEmptyRackMap(t *testing.T) {

	brokerReplicaCount := map[int32]int{
		1: 2,
		2: 3,
		3: 1,
	}

	replicas, err := getTargetReplicas([]int32{1, 2, 3}, brokerReplicaCount, 1, map[int32]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replicas) != 1 || replicas[0] != 3 {
		t.Fatalf("expected [3], got %v", replicas)
	}
}
