package deletion_test

import (
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/testutil"
)

func TestDeleteRecordsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "delete-records-", "--partitions", "2")

	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "a", 0, 0)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "b", 0, 1)
	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "c", 1, 0)
	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "d", 1, 1)
	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "e", 1, 2)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	// check initial messages
	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--print-keys", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(messages) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(messages))
	}

	// delete records
	if _, err := kafkaCtl.Execute("delete", "records", topicName, "--offset", "0=1", "--offset", "1=2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// check messages
	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--print-keys", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	messages = strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	testutil.AssertEquals(t, "key-1#b", messages[0])
	testutil.AssertEquals(t, "key-2#e", messages[1])
}

func TestDeleteRecordsAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "delete-complete-"

	topicName1 := testutil.CreateTopic(t, prefix+"a")
	topicName2 := testutil.CreateTopic(t, prefix+"b")
	topicName3 := testutil.CreateTopic(t, prefix+"c")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "delete", "records", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, topicName1, outputLines)
	testutil.AssertContains(t, topicName2, outputLines)
	testutil.AssertContains(t, topicName3, outputLines)
}
