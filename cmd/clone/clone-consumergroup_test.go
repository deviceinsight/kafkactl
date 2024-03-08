package clone_test

import (
	"fmt"
	"testing"

	"github.com/deviceinsight/kafkactl/internal/testutil"
)

func TestCloneConsumerGroupIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topic1 := testutil.CreateTopic(t, "topic1")
	topic2 := testutil.CreateTopic(t, "topic2")
	testutil.ProduceMessage(t, topic1, "test-key-1", "test-value-1", 0, 0)
	testutil.ProduceMessage(t, topic1, "test-key-1", "test-value-1", 0, 1)
	testutil.ProduceMessage(t, topic1, "test-key-2", "test-value-2a", 0, 2)
	testutil.ProduceMessage(t, topic2, "test-key-3", "test-value-3", 0, 0)
	testutil.ProduceMessage(t, topic2, "test-key-3", "test-value-3", 0, 1)

	srcGroup := testutil.CreateConsumerGroup(t, "srcGroup", topic1, topic2)
	targetGroup := testutil.GetPrefixedName("targetGroup")

	if _, err := kafkaCtl.Execute("clone", "cg", srcGroup, targetGroup); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("consumer-group %s cloned to %s", srcGroup, targetGroup), kafkaCtl.GetStdOut())

	testutil.VerifyConsumerGroupOffset(t, targetGroup, topic1, 3)
	testutil.VerifyConsumerGroupOffset(t, targetGroup, topic2, 2)
}

func TestCloneNonExistingConsumerGroupIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	srcGroup := testutil.GetPrefixedName("src-group")
	dstGroup := testutil.GetPrefixedName("dst-group")

	if _, err := kafkaCtl.Execute("clone", "cg", srcGroup, dstGroup); err != nil {
		testutil.AssertErrorContains(t, fmt.Sprintf("consumerGroup '%s' does not contain offsets", srcGroup), err)
	} else {
		t.Fatalf("Expected clone operation to fail")
	}
}

func TestCloneToExistingConsumerGroupIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topic := testutil.CreateTopic(t, "topic")

	srcGroup := testutil.CreateConsumerGroup(t, "src-group", topic)
	dstGroup := testutil.CreateConsumerGroup(t, "dst-group", topic)

	if _, err := kafkaCtl.Execute("clone", "cg", srcGroup, dstGroup); err != nil {
		testutil.AssertErrorContains(t, fmt.Sprintf("consumerGroup '%s' contains offsets", dstGroup), err)
	} else {
		t.Fatalf("Expected clone operation to fail")
	}
}
