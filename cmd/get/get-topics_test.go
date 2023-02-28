package get_test

import (
	"fmt"
	"testing"

	"github.com/deviceinsight/kafkactl/testutil"
)

func TestGetTopicsShowsMinReplicationFactorOfPartitionsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "get-topics", "--partitions", "2", "--replication-factor", "3")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "partition", topicName, "0", "--replicas", "101,102"); err != nil {
		testutil.AssertErrorContains(t, "version of API is not supported", err)
		return
	}

	if _, err := kafkaCtl.Execute("get", "topics"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := kafkaCtl.GetStdOutLines()

	testutil.AssertContains(t, "TOPIC|PARTITIONS|REPLICATION FACTOR", outputLines)
	testutil.AssertContains(t, fmt.Sprintf("%s|2|2", topicName), outputLines)
}

func TestGetTopicsShowsReplicationFactorOfTopicsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicA := testutil.CreateTopic(t, "get-topics", "--replication-factor", "1")
	topicB := testutil.CreateTopic(t, "get-topics", "--replication-factor", "2")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("get", "topics"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := kafkaCtl.GetStdOutLines()

	testutil.AssertContains(t, "TOPIC|PARTITIONS|REPLICATION FACTOR", outputLines)
	testutil.AssertContains(t, fmt.Sprintf("%s|1|1", topicA), outputLines)
	testutil.AssertContains(t, fmt.Sprintf("%s|1|2", topicB), outputLines)
}
