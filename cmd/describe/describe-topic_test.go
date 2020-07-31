package describe_test

import (
	"github.com/deviceinsight/kafkactl/test_util"
	"strings"
	"testing"
)

func TestDescribeTopicAutoCompletionIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	prefix := "describe-t-complete-"

	topicName1 := test_util.CreateTopic(t, prefix+"a")
	topicName2 := test_util.CreateTopic(t, prefix+"b")
	topicName3 := test_util.CreateTopic(t, prefix+"c")

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "describe", "topic", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, topicName1, outputLines)
	test_util.AssertContains(t, topicName2, outputLines)
	test_util.AssertContains(t, topicName3, outputLines)
}
