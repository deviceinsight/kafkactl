package reset_test

import (
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/testutil"
)

func TestResetCGOAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "reset-cgo-completion")

	prefix := "reset-cgo-complete-"

	group1 := testutil.CreateConsumerGroup(t, topicName, prefix+"a")
	group2 := testutil.CreateConsumerGroup(t, topicName, prefix+"b")
	group3 := testutil.CreateConsumerGroup(t, topicName, prefix+"c")

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
