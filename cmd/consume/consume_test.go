package consume_test

import (
	"fmt"
	"github.com/deviceinsight/kafkactl/test_util"
	"strings"
	"testing"
)

func TestConsumeWithKeyAndValueIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName := test_util.CreateTopic(t, "consume-topic")

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

func TestConsumeWithKeyAndValueAsBase64Integration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName := test_util.CreateTopic(t, "consume-topic")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute(
		"consume",
		topicName,
		"--from-beginning", "--exit", "--print-keys", "--key-encoding=base64", "--value-encoding=base64"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "dGVzdC1rZXk=#dGVzdC12YWx1ZQ==", kafkaCtl.GetStdOut())
}

func TestConsumeWithKeyAndValueAsHexIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName := test_util.CreateTopic(t, "consume-topic")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute(
		"consume",
		topicName,
		"--from-beginning", "--exit", "--print-keys", "--key-encoding=hex", "--value-encoding=hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "746573742d6b6579#746573742d76616c7565", kafkaCtl.GetStdOut())
}

func TestConsumeWithKeyAndValueAutoDetectBinaryValueIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName := test_util.CreateTopic(t, "consume-topic")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName,
		"--key", "test-key",
		"--value", "0000017373be345c", "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute(
		"consume",
		topicName,
		"--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "test-key#AAABc3O+NFw=", kafkaCtl.GetStdOut())
}

func TestConsumeAutoCompletionIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	prefix := "consume-complete-"

	topicName1 := test_util.CreateTopic(t, prefix+"a")
	topicName2 := test_util.CreateTopic(t, prefix+"b")
	topicName3 := test_util.CreateTopic(t, prefix+"c")

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "consume", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, topicName1, outputLines)
	test_util.AssertContains(t, topicName2, outputLines)
	test_util.AssertContains(t, topicName3, outputLines)
}

func TestAvroDeserializationErrorHandlingIntegration(t *testing.T) {

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

	topicName := test_util.CreateAvroTopic(t, "avro-topic", "", valueSchema)

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	// produce valid avro message
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value, "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	// produce message that cannot be deserialized
	test_util.SwitchContext("no-avro")

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "no-avro"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=1)", kafkaCtl.GetStdOut())

	test_util.SwitchContext("default")

	// produce another valid avro message
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value, "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=2)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit"); err != nil {
		test_util.AssertEquals(t, fmt.Sprintf("%s\n%s", value, value), kafkaCtl.GetStdOut())
		test_util.AssertErrorContains(t, "failed to find avro schema", err)
	} else {
		t.Fatalf("expected consumer to fail")
	}
}
