package produce_test

import (
	"fmt"
	"github.com/deviceinsight/kafkactl/test_util"
	"testing"
)

func TestProduceWithKeyAndValueIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

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

	test_util.StartIntegrationTest(t)

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

	test_util.StartIntegrationTest(t)

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

func TestProduceTombstoneFromFileIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName := test_util.CreateTopic(t, "produce-topic")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "-f", "../../test_resources/test-produce-tombstone.txt", "-S", "#"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "1 messages produced", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "ID123#", kafkaCtl.GetStdOut())
}

func TestProduceFromBase64Integration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName := test_util.CreateTopic(t, "produce-topic")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName,
		"--key", "dGVzdC1rZXk=", "--key-encoding", "base64",
		"--value", "dGVzdC12YWx1ZQ==", "--value-encoding", "base64"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "test-key#test-value", kafkaCtl.GetStdOut())
}

func TestProduceFromHexIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName := test_util.CreateTopic(t, "produce-topic")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName,
		"--key", "test-key",
		"--value", "0000000000000000", "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys", "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "test-key#0000000000000000", kafkaCtl.GetStdOut())
}
