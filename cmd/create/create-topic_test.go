package create_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestCreateTopicWithoutFlagsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.GetPrefixedName("new-topic")

	if _, err := kafkaCtl.Execute("create", "topic", topicName); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("topic created: %s", topicName), kafkaCtl.GetStdOut())

	describeTopic(t, kafkaCtl, topicName)
	stdOut := testutil.WithoutBrokerReferences(kafkaCtl.GetStdOut())

	expected := `
name: %s
replicationFactor: 1
partitions:
- id: 0
  oldestOffset: 0
  newestOffset: 0
  leader: any-broker
  replicas: [any-broker-id]
  inSyncReplicas: [any-broker-id]`

	testutil.AssertEquals(t, fmt.Sprintf(expected, topicName), stdOut)

}

func TestCreateTopicWithTwoPartitionsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.GetPrefixedName("new-topic")

	if _, err := kafkaCtl.Execute("create", "topic", topicName, "--partitions", "2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("topic created: %s", topicName), kafkaCtl.GetStdOut())

	describeTopic(t, kafkaCtl, topicName)
	stdOut := testutil.WithoutBrokerReferences(kafkaCtl.GetStdOut())

	expected := `
name: %s
replicationFactor: 1
partitions:
- id: 0
  oldestOffset: 0
  newestOffset: 0
  leader: any-broker
  replicas: [any-broker-id]
  inSyncReplicas: [any-broker-id]
- id: 1
  oldestOffset: 0
  newestOffset: 0
  leader: any-broker
  replicas: [any-broker-id]
  inSyncReplicas: [any-broker-id]`

	testutil.AssertEquals(t, fmt.Sprintf(expected, topicName), stdOut)

}

func TestCreateTopicWithReplicationFactorIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.GetPrefixedName("new-topic")

	if _, err := kafkaCtl.Execute("create", "topic", topicName, "-r", "3"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("topic created: %s", topicName), kafkaCtl.GetStdOut())

	describeTopic(t, kafkaCtl, topicName)
	stdOut := testutil.WithoutBrokerReferences(kafkaCtl.GetStdOut())

	expected := `
name: %s
replicationFactor: 3
partitions:
- id: 0
  oldestOffset: 0
  newestOffset: 0
  leader: any-broker
  replicas: [any-broker-id, any-broker-id, any-broker-id]
  inSyncReplicas: [any-broker-id, any-broker-id, any-broker-id]`

	testutil.AssertEquals(t, fmt.Sprintf(expected, topicName), stdOut)
}

func TestCreateTopicWithConfigFileIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.GetPrefixedName("new-topic")
	configFile := fmt.Sprintf(`
name: %s
partitions:
- id: 0
  oldestOffset: 0
  newestOffset: 290
  leader: kafka:9092
  replicas: [1]
  inSyncReplicas: [1]
- id: 1
  oldestOffset: 0
  newestOffset: 258
  leader: kafka:9092
  replicas: [1]
  inSyncReplicas: [1]
- id: 2
  oldestOffset: 0
  newestOffset: 290
  leader: kafka:9092
  replicas: [1]
  inSyncReplicas: [1]
configs:
- name: cleanup.policy
  value: compact
- name: delete.retention.ms
  value: "0"
- name: max.message.bytes
  value: "10485880"
- name: min.cleanable.dirty.ratio
  value: "1.0E-4"
- name: segment.ms
  value: "100"
`, topicName)

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.yaml", topicName))
	if err != nil {
		t.Fatalf("could not create temp file with topic config: %s", err)
	}
	defer tmpFile.Close()
	if _, err := tmpFile.WriteString(configFile); err != nil {
		t.Fatalf("could not write temp config file: %s", err)
	}

	if _, err := kafkaCtl.Execute("create", "topic", topicName, "-f", tmpFile.Name()); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("topic created: %s", topicName), kafkaCtl.GetStdOut())

	describeTopic(t, kafkaCtl, topicName)
	stdOut := testutil.WithoutBrokerReferences(kafkaCtl.GetStdOut())

	expected := `
name: %s
replicationFactor: 1
partitions:
- id: 0
  oldestOffset: 0
  newestOffset: 0
  leader: any-broker
  replicas: [any-broker-id]
  inSyncReplicas: [any-broker-id]
- id: 1
  oldestOffset: 0
  newestOffset: 0
  leader: any-broker
  replicas: [any-broker-id]
  inSyncReplicas: [any-broker-id]
- id: 2
  oldestOffset: 0
  newestOffset: 0
  leader: any-broker
  replicas: [any-broker-id]
  inSyncReplicas: [any-broker-id]
configs:
- name: cleanup.policy
  value: compact
- name: delete.retention.ms
  value: "0"
- name: max.message.bytes
  value: "10485880"
- name: min.cleanable.dirty.ratio
  value: "1.0E-4"
- name: segment.ms
  value: "100"`

	testutil.AssertEquals(t, fmt.Sprintf(expected, topicName), stdOut)
}

func describeTopic(t *testing.T, kafkaCtl testutil.KafkaCtlTestCommand, topicName string) {
	describeTopic := func(_ uint) error {
		_, err := kafkaCtl.Execute("describe", "topic", topicName, "-o", "yaml")
		return err
	}

	err := retry.Retry(
		describeTopic,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("failed to execute describe topic command: %v", err)
	}
}
