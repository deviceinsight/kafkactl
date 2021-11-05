package create_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/testutil"
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
partitions:
- id: 0
  oldestOffset: 0
  newestOffset: 0
  leader: any-broker
  replicas: [any-broker-id, any-broker-id, any-broker-id]
  inSyncReplicas: [any-broker-id, any-broker-id, any-broker-id]`

	testutil.AssertEquals(t, fmt.Sprintf(expected, topicName), stdOut)
}

func describeTopic(t *testing.T, kafkaCtl testutil.KafkaCtlTestCommand, topicName string) {
	describeTopic := func(attempt uint) error {
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
