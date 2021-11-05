package consume_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/testutil"
)

func TestConsumeWithKeyAndValueIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "test-key#test-value", kafkaCtl.GetStdOut())
}

func TestConsumeWithKeyAndValueAsBase64Integration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute(
		"consume",
		topicName,
		"--from-beginning", "--exit", "--print-keys", "--key-encoding=base64", "--value-encoding=base64"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "dGVzdC1rZXk=#dGVzdC12YWx1ZQ==", kafkaCtl.GetStdOut())
}

func TestConsumeWithKeyAndValueAsHexIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute(
		"consume",
		topicName,
		"--from-beginning", "--exit", "--print-keys", "--key-encoding=hex", "--value-encoding=hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "746573742d6b6579#746573742d76616c7565", kafkaCtl.GetStdOut())
}

func TestConsumeWithKeyAndValueAutoDetectBinaryValueIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName,
		"--key", "test-key",
		"--value", "0000017373be345c", "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute(
		"consume",
		topicName,
		"--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "test-key#AAABc3O+NFw=", kafkaCtl.GetStdOut())
}

func TestConsumeAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "consume-complete-"

	topicName1 := testutil.CreateTopic(t, prefix+"a")
	topicName2 := testutil.CreateTopic(t, prefix+"b")
	topicName3 := testutil.CreateTopic(t, prefix+"c")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "consume", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, topicName1, outputLines)
	testutil.AssertContains(t, topicName2, outputLines)
	testutil.AssertContains(t, topicName3, outputLines)
}

func TestAvroDeserializationErrorHandlingIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

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

	topicName := testutil.CreateAvroTopic(t, "avro-topic", "", valueSchema)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	// produce valid avro message
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value, "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	// produce message that cannot be deserialized
	testutil.SwitchContext("no-avro")

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "no-avro"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=1)", kafkaCtl.GetStdOut())

	testutil.SwitchContext("default")

	// produce another valid avro message
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value, "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=2)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit"); err != nil {
		testutil.AssertErrorContains(t, "failed to find avro schema", err)
		testutil.AssertEquals(t, fmt.Sprintf("%s\n%s", value, value), kafkaCtl.GetStdOut())
	} else {
		t.Fatalf("expected consumer to fail")
	}
}
