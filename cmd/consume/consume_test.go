package consume_test

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/deviceinsight/kafkactl/internal/helpers/protobuf"

	"github.com/deviceinsight/kafkactl/testutil"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func TestProtobufConsumeProtoFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "proto-file")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	protoPath := filepath.Join(testutil.RootDir, "testutil", "testdata")
	now := time.Date(2021, time.December, 1, 14, 10, 12, 0, time.UTC)
	pbMessageDesc := protobuf.ResolveMessageType(protobuf.SearchContext{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicMessage")
	pbMessage := dynamic.NewMessage(pbMessageDesc)
	pbMessage.SetFieldByNumber(1, timestamppb.New(now))
	pbMessage.SetFieldByNumber(2, int64(1))

	value, err := pbMessage.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal proto message: %s", err)
	}

	// produce valid pb message
	if _, err := kafkaCtl.Execute("produce", pbTopic, "--key", "test-key", "--value", hex.EncodeToString(value), "--value-encoding", "hex", "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--proto-import-path", protoPath, "--proto-file", "msg.proto", "--value-proto-type", "TopicMessage"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`, kafkaCtl.GetStdOut())
}

func TestProtobufConsumeProtosetFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "proto-file")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	protoPath := filepath.Join(testutil.RootDir, "testutil", "testdata", "msg.protoset")
	now := time.Date(2021, time.December, 1, 14, 10, 12, 0, time.UTC)
	pbMessageDesc := protobuf.ResolveMessageType(protobuf.SearchContext{
		ProtosetFiles: []string{protoPath},
	}, "TopicMessage")
	pbMessage := dynamic.NewMessage(pbMessageDesc)
	pbMessage.SetFieldByNumber(1, timestamppb.New(now))
	pbMessage.SetFieldByNumber(2, int64(1))

	value, err := pbMessage.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal proto message: %s", err)
	}

	// produce valid pb message
	if _, err := kafkaCtl.Execute("produce", pbTopic, "--key", "test-key", "--value", hex.EncodeToString(value), "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--protoset-file", protoPath, "--value-proto-type", "TopicMessage"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`, kafkaCtl.GetStdOut())
}

func TestProtobufConsumeProtoFileErrNoMessageIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "proto-file")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	protoPath := filepath.Join(testutil.RootDir, "testutil", "testdata", "msg.protoset")

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--proto-import-path", protoPath, "--proto-file", "msg.proto", "--value-proto-type", "NonExisting"); err != nil {
		testutil.AssertErrorContains(t, "not found in provided files", err)
	} else {
		t.Fatal("Expected consumer to fail")
	}
}

func TestProtobufConsumeProtoFileErrDecodeIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "proto-file")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	protoPath := filepath.Join(testutil.RootDir, "testutil", "testdata")

	// produce invalid pb message
	if _, err := kafkaCtl.Execute("produce", pbTopic, "--key", "test-key", "--value", "nonpb"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--proto-import-path", protoPath, "--proto-file", "msg.proto", "--value-proto-type", "TopicMessage"); err != nil {
		testutil.AssertErrorContains(t, "value decode failed: proto: bad wiretype", err)
	} else {
		t.Fatal("Expected consumer to fail")
	}
}
