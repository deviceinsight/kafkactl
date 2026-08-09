package produce_test

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/riferrei/srclient"

	produceCmd "github.com/deviceinsight/kafkactl/v5/cmd/produce"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestProduceRejectsEmptyAvroSchemaFile(t *testing.T) {
	cmd := produceCmd.NewProduceCmd()
	cmd.SetArgs([]string{"events", "--avro-schema-file=", "--value", `{"id":"123"}`})
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected --avro-schema-file= to return an error, got nil")
	}
	const wantError = "parameter --avro-schema-file must not be empty"
	if err.Error() != wantError {
		t.Fatalf("expected error %q, got %q", wantError, err)
	}
}

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

	topicName := testutil.CreateTopicWithSchema(t, "produce-topic", "", valueSchema, srclient.Avro)

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

func TestProduceAvroMessageOmitDefaultValueIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	valueSchema := `{
	  "name": "CreateUserProfileWallet",
	  "namespace": "Messaging.Contracts.WalletManager.Commands",
	  "type": "record",
	  "fields": [
		{ "name": "CurrencyCode", "type": "string" },
		{ "name": "ExpiresOn", "type": ["null", "string"], "default": null}
	  ]
	}`
	value := `{
	 "CurrencyCode": "EUR"
	}`

	topicName := testutil.CreateTopicWithSchema(t, "produce-avro-topic", "", valueSchema, srclient.Avro)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	stdout := kafkaCtl.GetStdOut()
	testutil.AssertContainSubstring(t, `"CurrencyCode":"EUR"`, stdout)
	testutil.AssertContainSubstring(t, `"ExpiresOn":null`, stdout)
}

func TestProduceAvroMessageWithUnionStandardJsonIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	valueSchema := `{
	  "name": "CreateUserProfileWallet",
	  "namespace": "Messaging.Contracts.WalletManager.Commands",
	  "type": "record",
	  "fields": [
		{ "name": "CurrencyCode", "type": "string" },
		{ "name": "ExpiresOn", "type": ["null", "string"], "default": null}
	  ]
	}`

	value := `{
	 "CurrencyCode": "EUR",
	 "ExpiresOn": "2022-12-12"
	}`

	topicName := testutil.CreateTopicWithSchema(t, "produce-topic", "", valueSchema, srclient.Avro)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	stdout := kafkaCtl.GetStdOut()
	testutil.AssertContainSubstring(t, `"CurrencyCode":"EUR"`, stdout)
	testutil.AssertContainSubstring(t, `"ExpiresOn":"2022-12-12"`, stdout)
}

func TestProduceRegistryProtobufMessageWithHeadersIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	valueSchema := `syntax = "proto3";
  package foo.bar;

  message Msg {
    string name = 1;
  }`
	value := `{"name":"Peter Mueller"}`

	topicName := testutil.CreateTopicWithSchema(t, "produce-protobuf-topic", "", valueSchema, srclient.Protobuf)

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

func TestProduceRegistryProtobufMessageOmitDefaultValueIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	valueSchema := `syntax = "proto3";
  package foo.bar;

  message Msg {
    string current_code = 1;
    string expires_on = 2;
  }`
	value := `{"currentCode":"EUR"}`

	topicName := testutil.CreateTopicWithSchema(t, "produce-protobuf-topic", "", valueSchema, srclient.Protobuf)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning",
		"--proto-marshal-option", "emitDefaultValues", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	stdout := kafkaCtl.GetStdOut()
	testutil.AssertContainSubstring(t, `"currentCode":"EUR"`, stdout)
	testutil.AssertContainSubstring(t, `"expiresOn":""`, stdout)
}

func TestProduceJsonMessageWithSchemaIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	valueSchema := `{
	  "$schema": "http://json-schema.org/draft-04/schema#",
	  "type": "object",
	  "properties": {
		"CurrencyCode": {
		  "type": "string"
		},
		"ExpiresOn": {
		  "type": "string",
		  "format": "date"
		}
	  },
	  "required": [
		"CurrencyCode",
		"ExpiresOn"
	  ]
	}`

	value := `{
	 "CurrencyCode": "EUR",
	 "ExpiresOn": "2022-12-12"
	}`

	topicName := testutil.CreateTopicWithSchema(t, "produce-topic", "", valueSchema, srclient.Json)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	stdout := kafkaCtl.GetStdOut()
	testutil.AssertContainSubstring(t, `"CurrencyCode": "EUR"`, stdout)
	testutil.AssertContainSubstring(t, `"ExpiresOn": "2022-12-12"`, stdout)
}

func TestProduceJsonMessageWithSchemaValidationErrorIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	valueSchema := `{
	  "$schema": "http://json-schema.org/draft-04/schema#",
	  "type": "object",
	  "properties": {
		"Name": {
		  "type": "string"
		},
		"Age": {
		  "type": "integer"
		}
	  },
	  "required": [
		"Name",
		"Age"
	  ]
	}`

	// Missing required field "Age"
	value := `{"Name": "Alice"}`

	topicName := testutil.CreateTopicWithSchema(t, "produce-topic", "", valueSchema, srclient.Json)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	_, err := kafkaCtl.Execute("produce", topicName, "--value", value)
	if err == nil {
		t.Fatalf("expected validation error but got none")
	}

	testutil.AssertErrorContains(t, "json data does not match schema", err)
}

func TestProduceJsonMessageWithSchemaAndPrintSchemaIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	valueSchema := `{
	  "$schema": "http://json-schema.org/draft-04/schema#",
	  "type": "object",
	  "properties": {
		"City": {
		  "type": "string"
		}
	  },
	  "required": [
		"City"
	  ]
	}`

	value := `{"City": "Berlin"}`

	topicName := testutil.CreateTopicWithSchema(t, "produce-topic", "", valueSchema, srclient.Json)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-schema", "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	stdout := kafkaCtl.GetStdOut()
	testutil.AssertContainSubstring(t, "City", stdout)
	testutil.AssertContainSubstring(t, "Berlin", stdout)
	testutil.AssertContainSubstring(t, "valueSchema:", stdout)
	testutil.AssertContainSubstring(t, "valueSchemaId:", stdout)
}

func TestProduceAvroMessageWithUnionAvroJsonIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	valueSchema := `{
	  "name": "CreateUserProfileWallet",
	  "namespace": "Messaging.Contracts.WalletManager.Commands",
	  "type": "record",
	  "fields": [
		{ "name": "CurrencyCode", "type": "string" },
		{ "name": "ExpiresOn", "type": ["null", "string"], "default": null}
	  ]
	}`

	value := `{
	 "CurrencyCode": "EUR",
	 "ExpiresOn": {"string": "2022-12-12"}
	}`

	if err := os.Setenv("AVRO_JSONCODEC", "avro"); err != nil {
		t.Fatalf("unable to set env variable: %v", err)
	}

	topicName := testutil.CreateTopicWithSchema(t, "produce-topic", "", valueSchema, srclient.Avro)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	stdout := kafkaCtl.GetStdOut()
	testutil.AssertContainSubstring(t, `"CurrencyCode":"EUR"`, stdout)
	testutil.AssertContainSubstring(t, `"ExpiresOn":{"string":"2022-12-12"}`, stdout)
}

func TestProduceNullValueStringTreatedAsTombstoneIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "produce-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "null"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "-o", "yaml", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	record := strings.ReplaceAll(kafkaCtl.GetStdOut(), "\n", " ")
	testutil.AssertContainSubstring(t, "value: null", record)
}

func TestProduceNullKeyStringTreatedAsNullKeyIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "produce-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "null", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "-o", "yaml", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	record := strings.ReplaceAll(kafkaCtl.GetStdOut(), "\n", " ")
	testutil.AssertContainNoSubstring(t, "key:", record)
	testutil.AssertContainSubstring(t, "value: test-value", record)
}

func TestProduceNullStringLiteralViaBase64Integration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "produce-topic")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	// base64("null") = "bnVsbA==" — use encoding to write the literal string "null"
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "bnVsbA==", "--value-encoding", "base64"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "test-key#null", kafkaCtl.GetStdOut())
}

func TestProduceRawAvroNullIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	schemaPath := filepath.Join(t.TempDir(), "null.avsc")
	if err := os.WriteFile(schemaPath, []byte(`"null"`), 0o600); err != nil {
		t.Fatalf("failed to write Avro null schema %q: %v", schemaPath, err)
	}

	topicName := testutil.CreateTopic(t, "raw-avro-null")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute(
		"produce", topicName,
		"--avro-schema-file", schemaPath,
		"--value", "null",
	); err != nil {
		t.Fatalf("failed to produce --value null with a raw Avro schema: %v", err)
	}

	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute(
		"produce", topicName,
		"--avro-schema-file", schemaPath,
		"--null-value",
	); err != nil {
		t.Fatalf("failed to produce --null-value with a raw Avro schema: %v", err)
	}

	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute(
		"produce", topicName,
		"--avro-schema-file", schemaPath,
		"--value", "bnVsbA==",
		"--value-encoding", "base64",
	); err != nil {
		t.Fatalf("failed to produce base64-encoded raw Avro null: %v", err)
	}

	client := testutil.CreateClient(t)
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("failed to close Kafka client: %v", err)
		}
	}()

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		t.Fatalf("failed to create Kafka consumer: %v", err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			t.Errorf("failed to close Kafka consumer: %v", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(topicName, 0, sarama.OffsetOldest)
	if err != nil {
		t.Fatalf("failed to consume topic %q partition 0: %v", topicName, err)
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			t.Errorf("failed to close Kafka partition consumer: %v", err)
		}
	}()

	expectedMessages := []struct {
		name            string
		expectTombstone bool
	}{
		{name: "--value null", expectTombstone: true},
		{name: "--null-value", expectTombstone: true},
		{name: "base64 Avro null"},
	}
	for offset, expected := range expectedMessages {
		select {
		case message := <-partitionConsumer.Messages():
			if message == nil {
				t.Fatalf("expected Kafka message at offset %d, got nil", offset)
			}
			if message.Offset != int64(offset) {
				t.Fatalf("expected message offset %d, got %d", offset, message.Offset)
			}
			if expected.expectTombstone {
				if message.Value != nil {
					t.Fatalf("expected %s to produce a tombstone, got a non-nil value of %d bytes (%x)", expected.name, len(message.Value), message.Value)
				}
				continue
			}
			if message.Value == nil {
				t.Fatal("expected base64-encoded Avro null to produce a non-nil zero-byte value, got a tombstone")
			}
			if len(message.Value) != 0 {
				t.Fatalf("expected base64-encoded Avro null value length 0, got %d bytes (%x)", len(message.Value), message.Value)
			}
		case consumerErr := <-partitionConsumer.Errors():
			t.Fatalf("failed to consume raw Avro null message at expected offset %d: %v", offset, consumerErr)
		case <-time.After(10 * time.Second):
			t.Fatalf("timed out after 10s waiting for raw Avro null message at expected offset %d", offset)
		}
	}
}

func TestProduceRawAvroProtobufKeyIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	const valueSchema = `{
  "type": "record",
  "name": "person",
  "fields": [
    {"name": "name", "type": "string"}
  ]
}`
	schemaPath := filepath.Join(t.TempDir(), "person.avsc")
	if err := os.WriteFile(schemaPath, []byte(valueSchema), 0o600); err != nil {
		t.Fatalf("failed to write raw Avro schema %q: %v", schemaPath, err)
	}

	topicName := testutil.CreateTopic(t, "raw-avro-protobuf-key")
	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")
	key := `{"fvalue":1.2}`
	value := `{"name":"Alice"}`

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute(
		"produce", topicName,
		"--key", key,
		"--key-proto-type", "TopicKey",
		"--proto-import-path", protoPath,
		"--proto-file", "msg.proto",
		"--avro-schema-file", schemaPath,
		"--value", value,
	); err != nil {
		t.Fatalf("failed to produce raw Avro value with Protobuf key: %v", err)
	}

	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute(
		"consume", topicName,
		"--from-beginning",
		"--exit",
		"--print-keys",
		"--key-encoding", "hex",
		"--value-encoding", "hex",
	); err != nil {
		t.Fatalf("failed to consume raw Avro value with Protobuf key: %v", err)
	}

	encodedParts := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "#")
	if len(encodedParts) != 2 {
		t.Fatalf("encoded key/value parts = %d, want 2 in output %q", len(encodedParts), kafkaCtl.GetStdOut())
	}
	rawKey, err := hex.DecodeString(encodedParts[0])
	if err != nil {
		t.Fatalf("failed to decode Protobuf key from hex %q: %v", encodedParts[0], err)
	}
	rawValue, err := hex.DecodeString(encodedParts[1])
	if err != nil {
		t.Fatalf("failed to decode raw Avro value from hex %q: %v", encodedParts[1], err)
	}

	keyMessage := dynamicpb.NewMessage(protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicKey"))
	if err = proto.Unmarshal(rawKey, keyMessage); err != nil {
		t.Fatalf("failed to decode Protobuf key: %v", err)
	}
	actualKey, err := marshalJSON(keyMessage)
	if err != nil {
		t.Fatalf("failed to convert Protobuf key to JSON: %v", err)
	}
	testutil.AssertEquals(t, key, string(actualKey))

	valueCodec, err := avro.NewMessageCodec(valueSchema, avro.Standard)
	if err != nil {
		t.Fatalf("failed to create raw Avro value codec: %v", err)
	}
	actualValue, err := valueCodec.DecodeBinary(rawValue)
	if err != nil {
		t.Fatalf("failed to decode raw Avro value: %v", err)
	}
	testutil.AssertEquals(t, value, string(actualValue))
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

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

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

	keyMessage := dynamicpb.NewMessage(protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicKey"))
	valueMessage := dynamicpb.NewMessage(protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicMessage"))

	if err = proto.Unmarshal(rawKey, keyMessage); err != nil {
		t.Fatalf("Unmarshal key failed: %s", err)
	}
	if err = proto.Unmarshal(rawValue, valueMessage); err != nil {
		t.Fatalf("Unmarshal value failed: %s", err)
	}

	actualKey, err := marshalJSON(keyMessage)
	if err != nil {
		t.Fatalf("Key to json failed: %s", err)
	}

	actualValue, err := marshalJSON(valueMessage)
	if err != nil {
		t.Fatalf("Value to json failed: %s", err)
	}

	testutil.AssertEquals(t, key, string(actualKey))
	testutil.AssertEquals(t, value, string(actualValue))
}

func TestProduceWithCSVFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)
	topic := testutil.CreateTopic(t, "produce-topic-csv")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	dataFilePath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

	if _, err := kafkaCtl.Execute("produce", topic, "--separator", ",",
		"--file", filepath.Join(dataFilePath, "msg.csv")); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "3 messages produced", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topic, "--from-beginning", "--print-keys", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "1#a\n2#b\n3#c", kafkaCtl.GetStdOut())
}

func TestProduceWithCSVFileWithTimestampsFirstColumnIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)
	topic := testutil.CreateTopic(t, "produce-topic-csv")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	dataFilePath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

	if _, err := kafkaCtl.Execute("produce", topic, "--separator", ",",
		"--file", filepath.Join(dataFilePath, "msg-ts1.csv")); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "3 messages produced", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topic, "--from-beginning", "--print-keys", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "1#a\n2#b\n3#c", kafkaCtl.GetStdOut())
}

func TestProduceWithCSVFileWithTimestampsSecondColumnIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)
	topic := testutil.CreateTopic(t, "produce-topic-csv")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	dataFilePath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

	if _, err := kafkaCtl.Execute("produce", topic, "--separator", ",",
		"--file", filepath.Join(dataFilePath, "msg-ts2.csv")); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "3 messages produced", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topic, "--from-beginning", "--print-keys", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "1#a\n2#b\n3#c", kafkaCtl.GetStdOut())
}

func TestProduceWithJSONFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)
	topic := testutil.CreateTopic(t, "produce-topic-json")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	dataFilePath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

	if _, err := kafkaCtl.Execute("produce", topic,
		"--file", filepath.Join(dataFilePath, "msg.json"),
		"--input-format", "json"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "6 messages produced", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topic, "--from-beginning", "--print-keys", "--print-headers", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	expectedMessages := []string{"a:b,c:1#1#a", "#2#b", "x:y#3#c", "##value-only", "#key-only#null", "##null"}
	testutil.AssertArraysEquals(t, expectedMessages, kafkaCtl.GetStdOutLines())
}

func TestProduceWithJSONFileBase64ValuesIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)
	topic := testutil.CreateTopic(t, "produce-topic-json-base64-values")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	dataFilePath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

	if _, err := kafkaCtl.Execute("produce", topic,
		"--file", filepath.Join(dataFilePath, "msg-base64.json"),
		"--value-encoding", "base64",
		"--input-format", "json"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "3 messages produced", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", topic, "--from-beginning", "--print-keys", "--value-encoding", "hex", "--exit"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "1#000000000001\n2#68656c6c6f\n3#6b61666b61", kafkaCtl.GetStdOut())
}

func TestProduceProtoFileWithOnlyKeyEncodedIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "produce-topic-pb")

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	key := `{"fvalue":1.2}`
	value := `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`

	if _, err := kafkaCtl.Execute("produce", pbTopic,
		"--key", key, "--key-proto-type", "TopicKey",
		"--value", value, "--proto-file", filepath.Join(protoPath, "msg.proto")); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	if _, err := kafkaCtl.Execute("consume", pbTopic, "--from-beginning", "--exit", "--print-keys", "--key-encoding", "hex"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	kv := strings.Split(kafkaCtl.GetStdOut(), "#")

	rawKey, err := hex.DecodeString(strings.TrimSpace(kv[0]))
	if err != nil {
		t.Fatalf("Failed to decode key: %s", err)
	}

	keyMessage := dynamicpb.NewMessage(protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicKey"))

	if err = proto.Unmarshal(rawKey, keyMessage); err != nil {
		t.Fatalf("Unmarshal key failed: %s", err)
	}

	actualKey, err := marshalJSON(keyMessage)
	if err != nil {
		t.Fatalf("Key to json failed: %s", err)
	}

	actualValue := strings.TrimSpace(kv[1])

	testutil.AssertEquals(t, key, string(actualKey))
	testutil.AssertEquals(t, value, actualValue)
}

func TestProduceProtoFileWithoutProtoImportPathIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "produce-topic-pb")

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	key := `{"fvalue":1.2}`
	value := `{"producedAt":"2021-12-01T14:10:12Z","num":"1"}`

	if _, err := kafkaCtl.Execute("produce", pbTopic,
		"--key", key, "--key-proto-type", "TopicKey",
		"--value", value, "--value-proto-type", "TopicMessage",
		"--proto-file", filepath.Join(protoPath, "msg.proto")); err != nil {
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

	keyMessage := dynamicpb.NewMessage(protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicKey"))
	valueMessage := dynamicpb.NewMessage(protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtoImportPaths: []string{protoPath},
		ProtoFiles:       []string{"msg.proto"},
	}, "TopicMessage"))

	if err = proto.Unmarshal(rawKey, keyMessage); err != nil {
		t.Fatalf("Unmarshal key failed: %s", err)
	}
	if err = proto.Unmarshal(rawValue, valueMessage); err != nil {
		t.Fatalf("Unmarshal value failed: %s", err)
	}

	actualKey, err := marshalJSON(keyMessage)
	if err != nil {
		t.Fatalf("Key to json failed: %s", err)
	}

	actualValue, err := marshalJSON(valueMessage)
	if err != nil {
		t.Fatalf("Value to json failed: %s", err)
	}

	testutil.AssertEquals(t, key, string(actualKey))
	testutil.AssertEquals(t, value, string(actualValue))
}

func TestProduceProtosetFileIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "produce-topic-pb")

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata", "msg.protoset")

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

	keyMessage := dynamicpb.NewMessage(protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtosetFiles: []string{protoPath},
	}, "TopicKey"))
	valueMessage := dynamicpb.NewMessage(protobuf.ResolveMessageType(internal.ProtobufConfig{
		ProtosetFiles: []string{protoPath},
	}, "TopicMessage"))

	if err = proto.Unmarshal(rawKey, keyMessage); err != nil {
		t.Fatalf("Unmarshal key failed: %s", err)
	}
	if err = proto.Unmarshal(rawValue, valueMessage); err != nil {
		t.Fatalf("Unmarshal value failed: %s", err)
	}

	actualKey, err := marshalJSON(keyMessage)
	if err != nil {
		t.Fatalf("Key to json failed: %s", err)
	}

	actualValue, err := marshalJSON(valueMessage)
	if err != nil {
		t.Fatalf("Value to json failed: %s", err)
	}

	testutil.AssertEquals(t, key, string(actualKey))
	testutil.AssertEquals(t, value, string(actualValue))
}

func TestProduceProtoFileBadJSONIntegration(t *testing.T) {
	testutil.StartIntegrationTest(t)

	pbTopic := testutil.CreateTopic(t, "produce-topic-pb")

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

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

	protoPath := filepath.Join(testutil.RootDir, "internal", "testutil", "testdata")

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

	file, err := os.CreateTemp(os.TempDir(), "long-message-")
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

	file, err := os.CreateTemp(os.TempDir(), "long-message-")
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

func marshalJSON(message *dynamicpb.Message) ([]byte, error) {
	jsonValue, err := protojson.MarshalOptions{Indent: ""}.Marshal(message)
	if err != nil {
		return nil, err
	}

	// this is needed to eliminate whitespace randomization
	// https://github.com/golang/protobuf/issues/1082
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, jsonValue); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
