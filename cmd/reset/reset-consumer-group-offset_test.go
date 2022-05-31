package reset_test

import (
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/testutil"
)

func TestResetCGOForSingleTopicIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "reset-cgo")

	prefix := "reset-cgo-"

	group1 := testutil.CreateConsumerGroup(t, prefix+"a", topicName)
	group2 := testutil.CreateConsumerGroup(t, prefix+"b", topicName)

	testutil.ProduceMessage(t, topicName, "test-key", "test-value1", 0, 0)
	testutil.ProduceMessage(t, topicName, "test-key", "test-value2", 0, 1)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("consume", topicName, "--group", group1, "--max-messages", "2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	consumed := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertContains(t, "test-value1", consumed)
	testutil.AssertContains(t, "test-value2", consumed)

	testutil.VerifyConsumerGroupOffset(t, group1, topicName, 2)
	testutil.VerifyConsumerGroupOffset(t, group2, topicName, 0)

	if _, err := kafkaCtl.Execute("reset", "consumer-group-offset", group1, "--topic", topicName, "--oldest", "--execute"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.VerifyConsumerGroupOffset(t, group1, topicName, 0)
	testutil.VerifyConsumerGroupOffset(t, group2, topicName, 0)

	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("reset", "consumer-group-offset", group2, "--topic", topicName, "--newest", "--execute"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.VerifyConsumerGroupOffset(t, group1, topicName, 0)
	testutil.VerifyConsumerGroupOffset(t, group2, topicName, 2)
}

func TestResetCGOForMultipleTopicsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicA := testutil.CreateTopic(t, "reset-cgo-a")
	topicB := testutil.CreateTopic(t, "reset-cgo-b")

	group := testutil.CreateConsumerGroup(t, "reset-cgo-multi", topicA, topicB)

	testutil.ProduceMessage(t, topicA, "test-key", "test-value1", 0, 0)
	testutil.ProduceMessage(t, topicA, "test-key", "test-value2", 0, 1)

	testutil.ProduceMessage(t, topicB, "test-key", "test-value1", 0, 0)
	testutil.ProduceMessage(t, topicB, "test-key", "test-value2", 0, 1)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("consume", topicA, "--group", group, "--max-messages", "2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	consumed := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertContains(t, "test-value1", consumed)
	testutil.AssertContains(t, "test-value2", consumed)

	testutil.VerifyConsumerGroupOffset(t, group, topicA, 2)
	testutil.VerifyConsumerGroupOffset(t, group, topicB, 0)

	if _, err := kafkaCtl.Execute("reset", "consumer-group-offset", group, "--topic", topicA, "--topic", topicB, "--oldest", "--execute"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.VerifyConsumerGroupOffset(t, group, topicA, 0)
	testutil.VerifyConsumerGroupOffset(t, group, topicB, 0)

	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("reset", "consumer-group-offset", group, "--topic", topicA, "--topic", topicB, "--newest", "--execute"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.VerifyConsumerGroupOffset(t, group, topicA, 2)
	testutil.VerifyConsumerGroupOffset(t, group, topicA, 2)
}

func TestResetCGOForAllTopicsInTheGroupIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicA := testutil.CreateTopic(t, "reset-cgo-a")
	topicB := testutil.CreateTopic(t, "reset-cgo-b")
	topicOther := testutil.CreateTopic(t, "reset-cgo-other")

	group := testutil.CreateConsumerGroup(t, "reset-cgo-all", topicA, topicB)

	testutil.ProduceMessage(t, topicA, "test-key", "test-value1", 0, 0)
	testutil.ProduceMessage(t, topicA, "test-key", "test-value2", 0, 1)

	testutil.ProduceMessage(t, topicB, "test-key", "test-value1", 0, 0)
	testutil.ProduceMessage(t, topicB, "test-key", "test-value2", 0, 1)

	testutil.ProduceMessage(t, topicOther, "test-key", "test-value1", 0, 0)
	testutil.ProduceMessage(t, topicOther, "test-key", "test-value2", 0, 1)

	testutil.VerifyConsumerGroupOffset(t, group, topicA, 0)
	testutil.VerifyConsumerGroupOffset(t, group, topicB, 0)
	testutil.VerifyTopicNotInConsumerGroup(t, group, topicOther)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("reset", "consumer-group-offset", group, "--all-topics", "--newest", "--execute"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.VerifyConsumerGroupOffset(t, group, topicA, 2)
	testutil.VerifyConsumerGroupOffset(t, group, topicB, 2)
	testutil.VerifyTopicNotInConsumerGroup(t, group, topicOther)
}

func TestResetCGOAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "reset-cgo-completion")

	prefix := "reset-cgo-complete-"

	group1 := testutil.CreateConsumerGroup(t, prefix+"a", topicName)
	group2 := testutil.CreateConsumerGroup(t, prefix+"b", topicName)
	group3 := testutil.CreateConsumerGroup(t, prefix+"c", topicName)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "reset", "consumer-group-offset", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, group1, outputLines)
	testutil.AssertContains(t, group2, outputLines)
	testutil.AssertContains(t, group3, outputLines)
}
