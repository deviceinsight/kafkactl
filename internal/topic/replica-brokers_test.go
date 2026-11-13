package topic

import (
	"net"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

func replicaTestClient(t *testing.T, includeUnreachable bool) sarama.Client {
	t.Helper()
	brokers := make([]*sarama.MockBroker, 3)
	metadata := sarama.NewMockMetadataResponse(t).SetController(1).
		SetLeader("example-events", 0, 1)
	for i := range brokers {
		brokers[i] = sarama.NewMockBroker(t, int32(i+1))
		t.Cleanup(brokers[i].Close)
		metadata.SetBroker(brokers[i].Addr(), int32(i+1))
	}
	if includeUnreachable {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		addr := listener.Addr().String()
		if err := listener.Close(); err != nil {
			t.Fatal(err)
		}
		metadata.SetBroker(addr, 4)
	}
	for _, broker := range brokers {
		broker.SetHandlerByMap(map[string]sarama.MockResponse{"MetadataRequest": metadata})
	}
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.ApiVersionsRequest = false
	config.Metadata.RefreshFrequency = 0
	config.Net.DialTimeout = time.Second
	config.Net.ReadTimeout = time.Second
	client, err := sarama.NewClient([]string{brokers[0].Addr()}, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	// Reading the existing partition opens its leader, not the other replicas.
	leader, err := client.Leader("example-events", 0)
	if err != nil {
		t.Fatal(err)
	}
	if connected, err := leader.Connected(); err != nil || !connected {
		t.Fatalf("leader did not connect: connected=%v, err=%v", connected, err)
	}
	return client
}

func TestAvailableReplicaBrokersOpensIdleConnections(t *testing.T) {
	client := replicaTestClient(t, false)
	for _, broker := range client.Brokers() {
		if broker.ID() != 1 {
			if connected, err := broker.Connected(); err != nil || connected {
				t.Fatalf("test needs unopened peer %d: connected=%v, err=%v", broker.ID(), connected, err)
			}
		}
	}
	available := availableReplicaBrokers(client)
	if len(available) != 3 {
		t.Fatalf("healthy but unopened peers were excluded: got %d available brokers, want 3", len(available))
	}
	counts := make(map[int32]int)
	for _, broker := range available {
		counts[broker.ID()] = 0
	}
	counts[1] = 1
	if replicas, err := getTargetReplicas([]int32{1}, counts, 3); err != nil || len(replicas) != 3 {
		t.Fatalf("cannot increase RF with healthy peers: replicas=%v, err=%v", replicas, err)
	}
}

func TestAvailableReplicaBrokersExcludesUnreachable(t *testing.T) {
	client := replicaTestClient(t, true)
	available := availableReplicaBrokers(client)
	if len(available) != 3 {
		t.Fatalf("got %d available brokers, want 3", len(available))
	}
	for _, broker := range available {
		if broker.ID() == 4 {
			t.Fatal("unreachable broker was accepted for replica assignment")
		}
	}
}
