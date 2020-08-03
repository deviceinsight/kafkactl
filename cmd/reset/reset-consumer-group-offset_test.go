package reset_test

import (
	"github.com/deviceinsight/kafkactl/test_util"
	"strings"
	"testing"
)

func TestResetCGOAutoCompletionIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName := test_util.CreateTopic(t, "reset-cgo-completion")

	prefix := "reset-cgo-complete-"

	group1 := test_util.CreateConsumerGroup(t, topicName, prefix+"a")
	group2 := test_util.CreateConsumerGroup(t, topicName, prefix+"b")
	group3 := test_util.CreateConsumerGroup(t, topicName, prefix+"c")

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "reset", "consumer-group-offset", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, group1, outputLines)
	test_util.AssertContains(t, group2, outputLines)
	test_util.AssertContains(t, group3, outputLines)
}
