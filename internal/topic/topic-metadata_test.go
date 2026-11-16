package topic

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

type partitionMetadataClient struct {
	sarama.Client
	counts        []int
	refreshCalls  int
	refreshErr    error
	partitionsErr error
	timeout       time.Duration
}

func (c *partitionMetadataClient) Config() *sarama.Config {
	config := sarama.NewConfig()
	config.Admin.Timeout = c.timeout
	return config
}

func (c *partitionMetadataClient) RefreshMetadata(...string) error {
	c.refreshCalls++
	return c.refreshErr
}

func (c *partitionMetadataClient) Partitions(string) ([]int32, error) {
	count := c.counts[min(c.refreshCalls-1, len(c.counts)-1)]
	return make([]int32, count), c.partitionsErr
}

func TestWaitForTopicPartitionsRefreshesStaleMetadata(t *testing.T) {
	client := &partitionMetadataClient{counts: []int{3, 6}, timeout: time.Second}
	if err := waitForTopicPartitions(client, "example-events", 6); err != nil {
		t.Fatal(err)
	}
	if client.refreshCalls != 2 {
		t.Fatalf("expected a second metadata refresh, got %d", client.refreshCalls)
	}
}

func TestWaitForTopicPartitionsTimesOut(t *testing.T) {
	client := &partitionMetadataClient{counts: []int{3}, timeout: time.Millisecond}
	err := waitForTopicPartitions(client, "example-events", 6)
	if err == nil || !strings.Contains(err.Error(), "expected at least 6") {
		t.Fatalf("expected a contextual metadata timeout, got %v", err)
	}
}

func TestWaitForTopicPartitionsReturnsReadErrors(t *testing.T) {
	for _, operation := range []string{"refresh", "partitions"} {
		t.Run(operation, func(t *testing.T) {
			cause := errors.New("synthetic metadata failure")
			client := &partitionMetadataClient{counts: []int{3}, timeout: time.Second}
			if operation == "refresh" {
				client.refreshErr = cause
			} else {
				client.partitionsErr = cause
			}
			if err := waitForTopicPartitions(client, "example-events", 6); !errors.Is(err, cause) {
				t.Fatalf("expected underlying read error, got %v", err)
			}
		})
	}
}
