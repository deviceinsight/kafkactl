package consume_test

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"github.com/riferrei/srclient"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
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
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}

	testutil.AssertEquals(t, "test-key-2#test-value-2a", messages[0])
	testutil.AssertEquals(t, "test-key-3#test-value-3", messages[1])
	testutil.AssertEquals(t, "test-key-2#test-value-2b", messages[2])
}

func TestConsumeFromTimestampIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic", "--partitions", "2")

	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "a", 0, 0)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "b", 0, 1)

	time.Sleep(1 * time.Millisecond) // need to have messaged produced at different milliseconds to have reproducible test
	t1 := time.Now().UnixMilli()

	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "c", 1, 0)
	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "d", 1, 1)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "e", 0, 2)
	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "f", 1, 2)

	time.Sleep(1 * time.Millisecond)
	t2 := time.Now().UnixMilli()

	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "g", 1, 3)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "h", 0, 3)

	// test --from-timestamp with --to-timestamp with formatted dates
	kafkaCtl := testutil.CreateKafkaCtlCommand()
	t1Formatted := time.UnixMilli(t1).Format("2006-01-02T15:04:05.000Z")
	t2Formatted := time.UnixMilli(t2).Format("2006-01-02T15:04:05.000Z")
	if _, err := kafkaCtl.Execute("consume", topicName, "--from-timestamp", t1Formatted, "--to-timestamp", t2Formatted); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertArraysEquals(t, []string{"c", "d", "e", "f"}, messages)

	// test --from-timestamp with --to-timestamp with unix epoch millis timestamps
	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("consume", topicName, "--from-timestamp", strconv.FormatInt(t1, 10), "--to-timestamp", strconv.FormatInt(t2, 10)); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	messages = strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertArraysEquals(t, []string{"c", "d", "e", "f"}, messages)

	// test --from-timestamp with --max-messages (--partitions present for reproducibility)
	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("consume", topicName, "--from-timestamp", strconv.FormatInt(t1, 10), "--max-messages", strconv.Itoa(2), "--partitions", strconv.Itoa(1)); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	messages = strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertArraysEquals(t, []string{"c", "d"}, messages)

	// test --from-timestamp with --exit
	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("consume", topicName, "--from-timestamp", strconv.FormatInt(t2, 10), "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	messages = strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertArraysEquals(t, []string{"g", "h"}, messages)
}

func TestConsumeRegistryProtobufWithNestedDependenciesIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	bazMsg := `syntax = "proto3";
  package baz;

  message Baz {
    string field = 1;
  }
  `

	barMsg := `syntax = "proto3";
  package bar;

  import "baz/protobuf/baz.proto";

  message Bar {
    baz.Baz bazField = 1;
  }
  `

	fooMsg := `syntax = "proto3";
  package foo;

  import "bar/protobuf/bar.proto";

  message Foo {
    bar.Bar barField = 1;
  }`

	value := `{"barField":{"bazField":{"field":"value"}}}`

	testutil.RegisterSchema(t, "baz", bazMsg, srclient.Protobuf)
	testutil.RegisterSchema(t, "bar", barMsg, srclient.Protobuf, srclient.Reference{Name: "baz/protobuf/baz.proto", Version: 1, Subject: "baz"})
	topicName := testutil.CreateTopic(t, "consume-topic")
	testutil.RegisterSchema(t, topicName+"-value", fooMsg, srclient.Protobuf, srclient.Reference{Name: "bar/protobuf/bar.proto", Version: 1, Subject: "bar"})

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("test-key#%s", value), kafkaCtl.GetStdOut())
}

func TestConsumeRegistryEmitDefaultValuesIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	singleMsg := `syntax = "proto3";
	
		message MyMessage {
		  message Metadata {
			string key = 1;
			string value = 2;
		  }
		  bool active = 3;
		  repeated Metadata metadata = 4;
		}
		`

	listMsg := `syntax = "proto3";

import "my-message.proto";

message MyList {
  repeated MyMessage messages = 3;
}
  `

	value := `{
  "messages": [
    {
      "active": true,
      "metadata": []
    },
    {
      "active": false,
      "metadata": []
    }
  ]
}`

	topicName := testutil.CreateTopic(t, "consume-topic")
	testutil.RegisterSchema(t, "my-message", singleMsg, srclient.Protobuf)
	testutil.RegisterSchema(t, topicName+"-value", listMsg, srclient.Protobuf, srclient.Reference{Name: "my-message.proto", Version: 1, Subject: "my-message"})

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys", "--proto-marshal-option", "emitdefaultvalues=true"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("test-key#%s", removeWhitespace(value)), removeWhitespace(kafkaCtl.GetStdOut()))
}

func TestConsumeRegistryProtobufWithWellKnowTypeIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	newFooMsg := `syntax = "proto3";
  package new.foo;

  import "google/protobuf/timestamp.proto";

  message Foo {
    google.protobuf.Timestamp field = 1;
  }
  `
	value := `{"field":"2025-06-07T11:11:11Z"}`

	testutil.RegisterSchema(t, "new.foo", newFooMsg, srclient.Protobuf)
	topicName := testutil.CreateTopic(t, "consume-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("test-key#%s", value), kafkaCtl.GetStdOut())
}

func TestConsumeRegistryProtobufWithNestedProtoIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	// example from https://docs.confluent.io/platform/current/schema-registry/fundamentals/serdes-develop/index.html#wire-format
	schema := `syntax = "proto3";
		package test.package;
		
		message MessageA {
			message MessageB {
				message MessageC {
					string fieldC = 1;
				}
			}
			message MessageD {
				string fieldD = 1;
			}
			message MessageE {
				message MessageF {
					string fieldF = 1;
				}
				message MessageG {
					string fieldG = 1;
				}
			}
		}
		message MessageH {
			message MessageI {
				string fieldI = 1;
			}
			MessageI fieldH = 1;
		}`

	testCases := []struct {
		name      string
		value     string
		protoType string
	}{
		{
			name:      "test_MessageH",
			value:     `{"fieldH":{"fieldI":"value"}}`,
			protoType: "test.package.MessageH",
		},
		{
			name:      "test_MessageI",
			value:     `{"fieldI":"value"}`,
			protoType: "test.package.MessageH.MessageI",
		},
		{
			name:      "test_MessageG",
			value:     `{"fieldG":"value"}`,
			protoType: "test.package.MessageA.MessageE.MessageG",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			topicName := testutil.CreateTopic(t, "consume-topic")
			testutil.RegisterSchema(t, topicName+"-value", schema, srclient.Protobuf)

			kafkaCtl := testutil.CreateKafkaCtlCommand()
			if _, err := kafkaCtl.Execute("produce", topicName, "--value-proto-type", tc.protoType, "--value", tc.value); err != nil {
				t.Fatalf("failed to execute command: %v", err)
			}
			testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

			if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit"); err != nil {
				t.Fatalf("failed to execute command: %v", err)
			}

			testutil.AssertEquals(t, tc.value, kafkaCtl.GetStdOut())
		})
	}
}

func TestConsumeToTimestampIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-topic", "--partitions", "2")

	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "a", 0, 0)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "b", 0, 1)

	time.Sleep(1 * time.Millisecond) // need to have messages produced at different milliseconds to have reproducible test
	t1 := time.Now().UnixMilli()

	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "c", 1, 0)
	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "d", 1, 1)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "e", 0, 2)
	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "f", 1, 2)

	time.Sleep(1 * time.Millisecond)
	t2 := time.Now().UnixMilli()

	testutil.ProduceMessageOnPartition(t, topicName, "key-2", "g", 1, 3)
	testutil.ProduceMessageOnPartition(t, topicName, "key-1", "h", 0, 3)

	// test --from-beginning with --to-timestamp
	kafkaCtl := testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--to-timestamp", strconv.FormatInt(t1, 10)); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertArraysEquals(t, []string{"a", "b"}, messages)

	// test --to-timestamp with --tail
	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("consume", topicName, "--to-timestamp", strconv.FormatInt(t2, 10), "--tail", strconv.Itoa(4)); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	messages = strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertArraysEquals(t, []string{"c", "d", "e", "f"}, messages)
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

	topicName := testutil.CreateTopicWithSchema(t, "avro-topic", "", valueSchema, srclient.Avro)

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
		t.Fatalf("failed to execute command: %v", err)
	} else {
		results := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
		testutil.AssertContains(t, value, results)
		testutil.AssertContains(t, "no-avro", results)
		testutil.AssertContains(t, value2, results)
	}

	kafkaCtl = testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--group", group, "--max-messages", "3"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	} else {
		results := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
		testutil.AssertContains(t, value, results)
		testutil.AssertContains(t, "no-avro", results)
		testutil.AssertContains(t, value2, results)
	}
}

func TestProtobufConsumeProtoFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "proto-file")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")
	now := time.Date(2021, time.December, 1, 14, 10, 12, 0, time.UTC)
	pbMessageDesc := protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicMessage")
	pbMessage := dynamicpb.NewMessage(pbMessageDesc)
	pbMessage.Set(pbMessageDesc.Fields().Get(0), protoreflect.ValueOfMessage(timestamppb.New(now).ProtoReflect()))
	pbMessage.Set(pbMessageDesc.Fields().Get(1), protoreflect.ValueOfInt64(1))

	value, err := proto.Marshal(pbMessage)
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

	// https://github.com/golang/protobuf/issues/1082
	output := strings.ReplaceAll(kafkaCtl.GetStdOut(), " ", "")

	testutil.AssertEquals(t, `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`, output)
}

func TestProtobufConsumeProtoFileWithoutProtoImportPathIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "proto-file")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")
	now := time.Date(2021, time.December, 1, 14, 10, 12, 0, time.UTC)
	pbMessageDesc := protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicMessage")
	pbMessage := dynamicpb.NewMessage(pbMessageDesc)
	pbMessage.Set(pbMessageDesc.Fields().Get(0), protoreflect.ValueOfMessage(timestamppb.New(now).ProtoReflect()))
	pbMessage.Set(pbMessageDesc.Fields().Get(1), protoreflect.ValueOfInt64(1))

	value, err := proto.Marshal(pbMessage)
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

	// https://github.com/golang/protobuf/issues/1082
	output := strings.ReplaceAll(kafkaCtl.GetStdOut(), " ", "")

	testutil.AssertEquals(t, `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`, output)
}

func TestConsumeTombstoneWithProtoFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "proto-file")
	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

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

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata", "msg.protoset")
	now := time.Date(2021, time.December, 1, 14, 10, 12, 0, time.UTC)
	pbMessageDesc := protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtosetFiles: []string{protoPath},
	}, "TopicMessage")
	pbMessage := dynamicpb.NewMessage(pbMessageDesc)
	pbMessage.Set(pbMessageDesc.Fields().Get(0), protoreflect.ValueOfMessage(timestamppb.New(now).ProtoReflect()))
	pbMessage.Set(pbMessageDesc.Fields().Get(1), protoreflect.ValueOfInt64(1))

	value, err := proto.Marshal(pbMessage)
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

	// https://github.com/golang/protobuf/issues/1082
	output := strings.ReplaceAll(kafkaCtl.GetStdOut(), " ", "")

	testutil.AssertEquals(t, `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`, output)
}

func TestProtobufConsumeProtoFileErrNoMessageIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "proto-file")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata", "msg.protoset")
	now := time.Date(2021, time.December, 1, 14, 10, 12, 0, time.UTC)
	pbMessageDesc := protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtosetFiles: []string{protoPath},
	}, "TopicMessage")
	pbMessage := dynamicpb.NewMessage(pbMessageDesc)
	pbMessage.Set(pbMessageDesc.Fields().Get(0), protoreflect.ValueOfMessage(timestamppb.New(now).ProtoReflect()))
	pbMessage.Set(pbMessageDesc.Fields().Get(1), protoreflect.ValueOfInt64(1))

	value, err := proto.Marshal(pbMessage)
	if err != nil {
		t.Fatalf("Failed to marshal proto message: %s", err)
	}

	// produce valid pb message
	if _, err := kafkaCtl.Execute("produce", pbTopic, "--key", "test-key", "--value", hex.EncodeToString(value), "--value-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

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

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

	// produce invalid pb message
	if _, err := kafkaCtl.Execute("produce", pbTopic, "--key", "test-key", "--value", "nonpb"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--proto-import-path", protoPath, "--proto-file", "msg.proto", "--value-proto-type", "TopicMessage"); err != nil {
		testutil.AssertErrorContains(t, "cannot parse invalid wire-format data", err)
	} else {
		t.Fatal("Expected consumer to fail")
	}
}

func TestConsumeGroupMaxMessagesDoNotOverConsumeIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)
	prefix := "consume-group-max-messages-test"
	topicName := testutil.CreateTopic(t, prefix+"topic")
	group := testutil.CreateConsumerGroup(t, prefix+"group", topicName)

	msgs := []string{"value1", "value2", "value3", "value4", "value5"}

	for i, msg := range msgs {
		testutil.ProduceMessage(t, topicName, "test-key", msg, 0, int64(i))
	}

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	for _, expectedMsg := range msgs {
		if _, err := kafkaCtl.Execute("consume", topicName, "--group", group, "--max-messages", "1"); err != nil {
			t.Fatalf("failed to execute command: %v", err)
		}
		results := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
		testutil.AssertContains(t, expectedMsg, results)
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

func TestConsumePartitionsK8sIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "k8s-mock")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	type testCases struct {
		description      string
		args             []string
		wantInKubectlCmd []string
	}

	for _, test := range []testCases{
		{
			description:      "single_partition_defined_with_space",
			args:             []string{"consume", "fake-topic", "--partitions", "5"},
			wantInKubectlCmd: []string{"--partitions=5"},
		},
		{
			description:      "single_partition_defined_with_equal",
			args:             []string{"consume", "fake-topic", "--partitions=5"},
			wantInKubectlCmd: []string{"--partitions=5"},
		},
		{
			description:      "multiple_partitions",
			args:             []string{"consume", "fake-topic", "--partitions", "5", "--partitions", "6"},
			wantInKubectlCmd: []string{"--partitions=5", "--partitions=6"},
		},
	} {
		t.Run(test.description, func(t *testing.T) {
			if _, err := kafkaCtl.Execute(test.args...); err != nil {
				t.Fatalf("failed to execute command: %v", err)
			}

			output := kafkaCtl.GetStdOut()

			for _, wanted := range test.wantInKubectlCmd {
				testutil.AssertContainSubstring(t, wanted, output)
			}
		})
	}
}

func removeWhitespace(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\v", "")
	s = strings.ReplaceAll(s, "\f", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func TestConsumeFilterByKeyIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-filter-topic")

	testutil.ProduceMessage(t, topicName, "user-123", "value1", 0, 0)
	testutil.ProduceMessage(t, topicName, "admin-456", "value2", 0, 1)
	testutil.ProduceMessage(t, topicName, "user-789", "value3", 0, 2)
	testutil.ProduceMessage(t, topicName, "guest-111", "value4", 0, 3)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--filter-key", "user-*", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d: %v", len(messages), messages)
	}

	testutil.AssertEquals(t, "user-123#value1", messages[0])
	testutil.AssertEquals(t, "user-789#value3", messages[1])
}

func TestConsumeFilterByKeyAlternativesIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-filter-topic")

	testutil.ProduceMessage(t, topicName, "user-123", "value1", 0, 0)
	testutil.ProduceMessage(t, topicName, "admin-456", "value2", 0, 1)
	testutil.ProduceMessage(t, topicName, "guest-789", "value3", 0, 2)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--filter-key", "{user,admin}-*", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d: %v", len(messages), messages)
	}

	testutil.AssertEquals(t, "user-123#value1", messages[0])
	testutil.AssertEquals(t, "admin-456#value2", messages[1])
}

func TestConsumeFilterByValueIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-filter-topic")

	testutil.ProduceMessage(t, topicName, "key1", "error: something failed", 0, 0)
	testutil.ProduceMessage(t, topicName, "key2", "success: operation completed", 0, 1)
	testutil.ProduceMessage(t, topicName, "key3", "error: connection timeout", 0, 2)
	testutil.ProduceMessage(t, topicName, "key4", "info: processing", 0, 3)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--filter-value", "error:*"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d: %v", len(messages), messages)
	}

	testutil.AssertEquals(t, "error: something failed", messages[0])
	testutil.AssertEquals(t, "error: connection timeout", messages[1])
}

func TestConsumeFilterByHeaderIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-filter-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "key1", "--value", "value1", "-H", "env:prod", "-H", "trace-id:abc-123"); err != nil {
		t.Fatalf("failed to produce message: %v", err)
	}

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "key2", "--value", "value2", "-H", "env:dev", "-H", "trace-id:xyz-456"); err != nil {
		t.Fatalf("failed to produce message: %v", err)
	}

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "key3", "--value", "value3", "-H", "env:prod", "-H", "trace-id:abc-789"); err != nil {
		t.Fatalf("failed to produce message: %v", err)
	}

	kafkaCtl = testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--filter-header", "trace-id=abc-*", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d: %v", len(messages), messages)
	}

	testutil.AssertEquals(t, "key1#value1", messages[0])
	testutil.AssertEquals(t, "key3#value3", messages[1])
}

func TestConsumeFilterANDLogicIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-filter-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "user-123", "--value", "error occurred", "-H", "env:prod"); err != nil {
		t.Fatalf("failed to produce message: %v", err)
	}

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "user-456", "--value", "success", "-H", "env:prod"); err != nil {
		t.Fatalf("failed to produce message: %v", err)
	}

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "admin-789", "--value", "error occurred", "-H", "env:prod"); err != nil {
		t.Fatalf("failed to produce message: %v", err)
	}

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "user-111", "--value", "error occurred", "-H", "env:dev"); err != nil {
		t.Fatalf("failed to produce message: %v", err)
	}

	kafkaCtl = testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--filter-key", "user-*", "--filter-value", "error*", "--filter-header", "env=prod", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d: %v", len(messages), messages)
	}

	testutil.AssertEquals(t, "user-123#error occurred", messages[0])
}

func TestConsumeFilterNoMatchIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "consume-filter-topic")

	testutil.ProduceMessage(t, topicName, "user-123", "value1", 0, 0)
	testutil.ProduceMessage(t, topicName, "admin-456", "value2", 0, 1)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--filter-key", "guest-*"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	output := strings.TrimSpace(kafkaCtl.GetStdOut())

	if output != "" {
		t.Fatalf("expected no output for non-matching filter, got: %q", output)
	}
}
