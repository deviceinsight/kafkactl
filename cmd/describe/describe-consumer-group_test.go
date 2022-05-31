package describe_test

import (
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/testutil"
)

func TestDescribeConsumerGroupTopicAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "describe-cg-complete-"

	topicName1 := testutil.CreateTopic(t, prefix+"a")
	topicName2 := testutil.CreateTopic(t, prefix+"b")
	topicName3 := testutil.CreateTopic(t, prefix+"c")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "describe", "consumer-group", "--topic", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, topicName1, outputLines)
	testutil.AssertContains(t, topicName2, outputLines)
	testutil.AssertContains(t, topicName3, outputLines)
}

func TestDescribeConsumerGroupCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "describe-cg-complete-"

	topicName := testutil.CreateTopic(t, prefix+"topic")

	group1 := testutil.CreateConsumerGroup(t, prefix+"a", topicName)
	group2 := testutil.CreateConsumerGroup(t, prefix+"b", topicName)
	group3 := testutil.CreateConsumerGroup(t, prefix+"c", topicName)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "describe", "consumer-group", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, group1, outputLines)
	testutil.AssertContains(t, group2, outputLines)
	testutil.AssertContains(t, group3, outputLines)
}
