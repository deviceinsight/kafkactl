package consume_test

import (
	"encoding/hex"
	"path/filepath"
	"strconv"
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

	testutil.ProduceMessage(t, topicName, "test-key", "test-value", 0, 0)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "test-key#test-value", kafkaCtl.GetStdOut())
}

func TestConsumeWithPartitionAndValueIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic", "--partitions", "2")

	testutil.ProduceMessage(t, topicName, "test-key", "test-value", 1, 0)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-partitions"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "1#test-value", kafkaCtl.GetStdOut())
}

func TestConsumeWithEmptyPartitionsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic", "--partitions", "10")

	testutil.ProduceMessage(t, topicName, "test-key", "test-value", 1, 0)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "test-key#test-value", kafkaCtl.GetStdOut())
}

func TestConsumeTailIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic", "--partitions", "10")

	testutil.ProduceMessage(t, topicName, "test-key-1", "test-value-1", 6, 0)
	testutil.ProduceMessage(t, topicName, "test-key-2", "test-value-2a", 2, 0)
	testutil.ProduceMessage(t, topicName, "test-key-3", "test-value-3", 5, 0)
	testutil.ProduceMessage(t, topicName, "test-key-2", "test-value-2b", 2, 1)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--tail", "3", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(messages) != 3 {
		t.Fatalf("expected 3 messages")
	}

	testutil.AssertEquals(t, "test-key-2#test-value-2a", messages[0])
	testutil.AssertEquals(t, "test-key-3#test-value-3", messages[1])
	testutil.AssertEquals(t, "test-key-2#test-value-2b", messages[2])
}

func TestConsumeFromTimestamp(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic", "--partitions", "2")

	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "a", 0, 0)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "b", 0, 1)

	time.Sleep(1 * time.Millisecond) // need to have messaged produced at different milliseconds to have reproductible test
	t1 := time.Now().UnixMilli()

	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "c", 1, 0)
	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "d", 1, 1)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "e", 0, 2)
	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "f", 1, 2)

	time.Sleep(1 * time.Millisecond)
	t2 := time.Now().UnixMilli()

	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "g", 1, 3)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "h", 0, 3)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-timestamp", strconv.FormatInt(t1, 10), "--to-timestamp", strconv.FormatInt(t2, 10)); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(messages) != 4 {
		t.Fatalf("expected 4 messages. Got %d : %s", len(messages), messages)
	}
	testutil.AssertContains(t, "c", messages)
	testutil.AssertContains(t, "d", messages)
	testutil.AssertContains(t, "e", messages)
	testutil.AssertContains(t, "f", messages)
}

func TestConsumeWithKeyAndValueAsBase64Integration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic")

	testutil.ProduceMessage(t, topicName, "test-key", "test-value", 0, 0)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

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

	testutil.ProduceMessage(t, topicName, "test-key", "test-value", 0, 0)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

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
	value2 := `{"name":"Peter Pan"}`

	topicName := testutil.CreateAvroTopic(t, "avro-topic", "", valueSchema)

	group := testutil.CreateConsumerGroup(t, "avro-topic-consumer-group", topicName)

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
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value2, "-H", "key1:value1", "-H", "key\\:2:value\\:2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=2)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit"); err != nil {
		testutil.AssertErrorContains(t, "failed to find avro schema", err)
		testutil.AssertEquals(t, value, kafkaCtl.GetStdOut())
	} else {
		t.Fatalf("expected consumer to fail")
	}

	kafkaCtl = testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--group", group, "--max-messages", "3"); err != nil {
		testutil.AssertErrorContains(t, "failed to find avro schema", err)
		testutil.AssertEquals(t, value, kafkaCtl.GetStdOut())
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
	if _, err := kafkaCtl.Execute("produce", pbTopic, "--key", "test-key", "--value", hex.EncodeToString(value), "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--proto-import-path", protoPath, "--proto-file", "msg.proto", "--value-proto-type", "TopicMessage"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`, kafkaCtl.GetStdOut())
}

func TestProtobufConsumeProtoFileWithoutProtoImportPathIntegration(t *testing.T) {
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
	if _, err := kafkaCtl.Execute("produce", pbTopic, "--key", "test-key", "--value", hex.EncodeToString(value), "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--proto-file", filepath.Join(protoPath, "msg.proto"), "--value-proto-type", "TopicMessage"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`, kafkaCtl.GetStdOut())
}

func TestConsumeTombstoneWithProtoFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "proto-file")
	protoPath := filepath.Join(testutil.RootDir, "testutil", "testdata")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", pbTopic, "--null-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "-o", "yaml", "--proto-import-path", protoPath, "--proto-file", "msg.proto", "--value-proto-type", "TopicMessage"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	record := strings.ReplaceAll(kafkaCtl.GetStdOut(), "\n", " ")
	testutil.AssertEquals(t, "partition: 0 offset: 0 value: null", record)
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

func TestConsumeGroupIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "consume-group-"

	topicName := testutil.CreateTopic(t, prefix+"topic")

	group1 := testutil.CreateConsumerGroup(t, prefix+"a", topicName)

	testutil.ProduceMessage(t, topicName, "test-key", "test-value1", 0, 0)

	group2 := testutil.CreateConsumerGroup(t, prefix+"b", topicName)

	testutil.ProduceMessage(t, topicName, "test-key", "test-value2", 0, 1)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--group", group1, "--max-messages", "2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	results := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertContains(t, "test-value1", results)
	testutil.AssertContains(t, "test-value2", results)

	if _, err := kafkaCtl.Execute("consume", topicName, "--group", group2, "--max-messages", "1"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	results = strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertContains(t, "test-value2", results)
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

func TestConsumeGroupCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "consume-group-complete-"

	topicName := testutil.CreateTopic(t, prefix+"topic")

	group1 := testutil.CreateConsumerGroup(t, prefix+"a", topicName)
	group2 := testutil.CreateConsumerGroup(t, prefix+"b", topicName)
	group3 := testutil.CreateConsumerGroup(t, prefix+"c", topicName)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "consume", topicName, "--group", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, group1, outputLines)
	testutil.AssertContains(t, group2, outputLines)
	testutil.AssertContains(t, group3, outputLines)
}
