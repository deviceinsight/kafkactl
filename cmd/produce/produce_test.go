package produce_test

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/internal/helpers/protobuf"

	"github.com/jhump/protoreflect/dynamic"

	"github.com/deviceinsight/kafkactl/testutil"
)

func TestProduceWithKeyAndValueIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "produce-topic")

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

func TestProduceMessageWithHeadersIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "produce-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value", "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys", "--print-headers"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "key1:value1,key\\:2:value\\:2#test-key#test-value", kafkaCtl.GetStdOut())
}

func TestProduceAvroMessageWithHeadersIntegration(t *testing.T) {

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

	topicName := testutil.CreateAvroTopic(t, "produce-topic", "", valueSchema)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value, "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys", "--print-headers"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("key1:value1,key\\:2:value\\:2#test-key#%s", value), kafkaCtl.GetStdOut())
}

func TestProduceTombstoneIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "produce-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--null-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "-o", "yaml", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	record := strings.ReplaceAll(kafkaCtl.GetStdOut(), "\n", " ")
	testutil.AssertEquals(t, "partition: 0 offset: 0 value: null", record)
}

func TestProduceFromBase64Integration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "produce-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName,
		"--key", "dGVzdC1rZXk=", "--key-encoding", "base64",
		"--value", "dGVzdC12YWx1ZQ==", "--value-encoding", "base64"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "test-key#test-value", kafkaCtl.GetStdOut())
}

func TestProduceFromHexIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "produce-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName,
		"--key", "test-key",
		"--value", "0000000000000000", "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys", "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "test-key#0000000000000000", kafkaCtl.GetStdOut())
}

func TestProduceAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "produce-complete-"

	topicName1 := testutil.CreateTopic(t, prefix+"a")
	topicName2 := testutil.CreateTopic(t, prefix+"b")
	topicName3 := testutil.CreateTopic(t, prefix+"c")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "produce", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, topicName1, outputLines)
	testutil.AssertContains(t, topicName2, outputLines)
	testutil.AssertContains(t, topicName3, outputLines)
}

func TestProduceProtoFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "produce-topic-pb")

	protoPath := filepath.Join(testutil.RootDir, "testutil", "testdata")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	key := `{"fvalue":1.2}`
	value := `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`

	if _, err := kafkaCtl.Execute("produce", pbTopic,
		"--key", key, "--key-proto-type", "TopicKey",
		"--value", value, "--value-proto-type", "TopicMessage",
		"--proto-import-path", protoPath, "--proto-file", "msg.proto"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--print-keys", "--key-encoding", "hex", "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	kv := strings.Split(kafkaCtl.GetStdOut(), "#")

	rawKey, err := hex.DecodeString(strings.TrimSpace(kv[0]))
	if err != nil {
		t.Fatalf("Failed to decode key: %s", err)
	}

	rawValue, err := hex.DecodeString(strings.TrimSpace(kv[1]))
	if err != nil {
		t.Fatalf("Failed to decode value: %s", err)
	}

	keyMessage := dynamic.NewMessage(protobuf.ResolveMessageType(protobuf.SearchContext{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicKey"))
	valueMessage := dynamic.NewMessage(protobuf.ResolveMessageType(protobuf.SearchContext{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicMessage"))

	if err = keyMessage.Unmarshal(rawKey); err != nil {
		t.Fatalf("Unmarshal key failed: %s", err)
	}
	if err = valueMessage.Unmarshal(rawValue); err != nil {
		t.Fatalf("Unmarshal value failed: %s", err)
	}

	actualKey, err := keyMessage.MarshalJSON()
	if err != nil {
		t.Fatalf("Key to json failed: %s", err)
	}

	actualValue, err := valueMessage.MarshalJSON()
	if err != nil {
		t.Fatalf("Value to json failed: %s", err)
	}

	testutil.AssertEquals(t, key, string(actualKey))
	testutil.AssertEquals(t, value, string(actualValue))
}

func TestProduceProtosetFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "produce-topic-pb")

	protoPath := filepath.Join(testutil.RootDir, "testutil", "testdata", "msg.protoset")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	key := `{"fvalue":1.2}`
	value := `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`

	if _, err := kafkaCtl.Execute("produce", pbTopic,
		"--key", key, "--key-proto-type", "TopicKey",
		"--value", value, "--value-proto-type", "TopicMessage",
		"--protoset-file", protoPath); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--print-keys", "--key-encoding", "hex", "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	kv := strings.Split(kafkaCtl.GetStdOut(), "#")

	rawKey, err := hex.DecodeString(strings.TrimSpace(kv[0]))
	if err != nil {
		t.Fatalf("Failed to decode key: %s", err)
	}

	rawValue, err := hex.DecodeString(strings.TrimSpace(kv[1]))
	if err != nil {
		t.Fatalf("Failed to decode value: %s", err)
	}

	keyMessage := dynamic.NewMessage(protobuf.ResolveMessageType(protobuf.SearchContext{
		ProtosetFiles: []string{protoPath},
	}, "TopicKey"))
	valueMessage := dynamic.NewMessage(protobuf.ResolveMessageType(protobuf.SearchContext{
		ProtosetFiles: []string{protoPath},
	}, "TopicMessage"))

	if err = keyMessage.Unmarshal(rawKey); err != nil {
		t.Fatalf("Unmarshal key failed: %s", err)
	}
	if err = valueMessage.Unmarshal(rawValue); err != nil {
		t.Fatalf("Unmarshal value failed: %s", err)
	}

	actualKey, err := keyMessage.MarshalJSON()
	if err != nil {
		t.Fatalf("Key to json failed: %s", err)
	}

	actualValue, err := valueMessage.MarshalJSON()
	if err != nil {
		t.Fatalf("Value to json failed: %s", err)
	}

	testutil.AssertEquals(t, key, string(actualKey))
	testutil.AssertEquals(t, value, string(actualValue))
}

func TestProduceProtoFileBadJSONIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "produce-topic-pb")

	protoPath := filepath.Join(testutil.RootDir, "testutil", "testdata")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	value := `{"producedAt":"2021-12-01T14:10:1`

	if _, err := kafkaCtl.Execute("produce", pbTopic,
		"--value", value, "--value-proto-type", "TopicMessage",
		"--proto-import-path", protoPath, "--proto-file", "msg.proto"); err != nil {
		testutil.AssertErrorContains(t, "invalid json", err)
	} else {
		t.Fatalf("Expected producer to fail")
	}
}

func TestProduceProtoFileErrNoMessageIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "produce-topic-pb")

	protoPath := filepath.Join(testutil.RootDir, "testutil", "testdata")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	value := `{"producedAt":"2021-12-01T14:10:1`

	if _, err := kafkaCtl.Execute("produce", pbTopic,
		"--value", value, "--value-proto-type", "unknown",
		"--proto-import-path", protoPath, "--proto-file", "msg.proto"); err != nil {
		testutil.AssertErrorContains(t, "not found in provided files", err)
	} else {
		t.Fatalf("Expected producer to fail")
	}
}

func TestProduceLongMessageSucceedsIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topic := testutil.CreateTopic(t, "produce-topic-long")

	file, err := ioutil.TempFile(os.TempDir(), "long-message-")
	if err != nil {
		t.Fatalf("unable to generate test file: %v", err)
	}
	defer os.Remove(file.Name())

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	data := make([]byte, bufio.MaxScanTokenSize)

	for i := range data {
		data[i] = 'K'
	}

	if err := os.WriteFile(file.Name(), data, 0644); err != nil {
		t.Fatalf("unable to write test file: %v", err)
	}

	if _, err := kafkaCtl.Execute("produce", topic, "--file", file.Name()); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "1 messages produced", kafkaCtl.GetStdOut())
}

func TestProduceLongMessageFailsIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topic := testutil.CreateTopic(t, "produce-topic-long")

	file, err := ioutil.TempFile(os.TempDir(), "long-message-")
	if err != nil {
		t.Fatalf("unable to generate test file: %v", err)
	}
	defer os.Remove(file.Name())

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	data := make([]byte, bufio.MaxScanTokenSize)

	for i := range data {
		data[i] = 'K'
	}

	if err := os.WriteFile(file.Name(), data, 0644); err != nil {
		t.Fatalf("unable to write test file: %v", err)
	}

	if _, err := kafkaCtl.Execute("produce", topic, "--max-message-bytes", strconv.Itoa(bufio.MaxScanTokenSize), "--file", file.Name()); err != nil {
		testutil.AssertErrorContains(t, "error reading input (try specifying --max-message-bytes when producing long messages)", err)
	} else {
		t.Fatalf("Expected producer to fail")
	}
}
