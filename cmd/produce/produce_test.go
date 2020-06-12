package produce_test

import (
	"github.com/deviceinsight/kafkactl/test_util"
	"testing"
)

func TestProduceWithKeyAndValueIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	topicName := test_util.CreateTopic(t, "produce-topic")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "test-key#test-value", kafkaCtl.GetStdOut())
}
