package produce_test

import (
	"fmt"
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

func TestProduceMessageWithHeadersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	topicName := test_util.CreateTopic(t, "produce-topic")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value", "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys", "--print-headers"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "key1:value1,key\\:2:value\\:2#test-key#test-value", kafkaCtl.GetStdOut())
}

func TestProduceAvroMessageWithHeadersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	valueSchema := `{
  "name": "person",
  "type": "record",
  "fields": [
	{
      "name": "name",
      "type": "string"
    }
  ]
}`
	value := `{"name":"Peter Mueller"}`

	topicName := test_util.CreateAvroTopic(t, "produce-topic", "", valueSchema)

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value, "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys", "--print-headers"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, fmt.Sprintf("key1:value1,key\\:2:value\\:2#test-key#%s", value), kafkaCtl.GetStdOut())
}
