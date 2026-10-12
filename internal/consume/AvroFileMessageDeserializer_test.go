package consume

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/linkedin/goavro/v2"
)

const testAvroFileSchema = `{"type":"record","name":"test","fields":[{"name":"id","type":"string"}]}`

func writeTestSchemaFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	schemaFile := filepath.Join(dir, "test.avsc")
	if err := os.WriteFile(schemaFile, []byte(testAvroFileSchema), 0o600); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}
	return schemaFile
}

func createTestDeserializer(t *testing.T, schemaFile, schemaHeader string) *AvroFileMessageDeserializer {
	t.Helper()
	deserializer, err := CreateAvroFileMessageDeserializer(schemaFile, schemaHeader, avro.NewSchemaSource(avro.Standard))
	if err != nil {
		t.Fatalf("failed to create deserializer: %v", err)
	}
	return deserializer
}

func TestAvroFileMessageDeserializer_CanDeserializeKeyAlwaysFalse(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, writeTestSchemaFile(t), "")

	if deserializer.CanDeserializeKey(&sarama.ConsumerMessage{}, Flags{}) {
		t.Fatal("expected CanDeserializeKey to always return false")
	}
}

func TestAvroFileMessageDeserializer_CanDeserializeValueWithSchemaFile(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, writeTestSchemaFile(t), "")

	if !deserializer.CanDeserializeValue(&sarama.ConsumerMessage{}, Flags{}) {
		t.Fatal("expected CanDeserializeValue to return true when schemaFile is set")
	}
}

func TestAvroFileMessageDeserializer_CanDeserializeValueWithHeaderPresent(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("ce_dataschema"), Value: []byte("http://example.com/schema.avsc")},
		},
	}

	if !deserializer.CanDeserializeValue(msg, Flags{}) {
		t.Fatal("expected CanDeserializeValue to return true when the configured header is present")
	}
}

func TestAvroFileMessageDeserializer_CanDeserializeValueWithHeaderAbsent(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	msg := &sarama.ConsumerMessage{}

	if deserializer.CanDeserializeValue(msg, Flags{}) {
		t.Fatal("expected CanDeserializeValue to return false when the configured header is absent")
	}
}

func TestAvroFileMessageDeserializer_CanDeserializeValueWithValueProtoTypeSet(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, writeTestSchemaFile(t), "")

	if deserializer.CanDeserializeValue(&sarama.ConsumerMessage{}, Flags{ValueProtoType: "com.example.Foo"}) {
		t.Fatal("expected CanDeserializeValue to return false when ValueProtoType is set")
	}
}

func TestAvroFileMessageDeserializer_DeserializeValueTombstone(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, writeTestSchemaFile(t), "")

	data, err := deserializer.DeserializeValue(&sarama.ConsumerMessage{Value: nil})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data.data != nil {
		t.Fatalf("expected nil data for tombstone record, got %v", data.data)
	}
}

func TestAvroFileMessageDeserializer_DeserializeValueFromSchemaFile(t *testing.T) {
	t.Parallel()

	schemaFile := writeTestSchemaFile(t)

	codec, err := goavro.NewCodecForStandardJSONFull(testAvroFileSchema)
	if err != nil {
		t.Fatalf("failed to create codec: %v", err)
	}

	native, _, err := codec.NativeFromTextual([]byte(`{"id":"abc"}`))
	if err != nil {
		t.Fatalf("failed to convert to native: %v", err)
	}

	binaryValue, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		t.Fatalf("failed to convert to binary: %v", err)
	}

	deserializer := createTestDeserializer(t, schemaFile, "")

	data, err := deserializer.DeserializeValue(&sarama.ConsumerMessage{Value: binaryValue})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `{"id":"abc"}`
	if string(data.data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data.data))
	}

	if !strings.Contains(data.schema, `"name":"test"`) {
		t.Fatalf("expected schema field to hold the schema text, not the file path, got: %s", data.schema)
	}
}

func TestAvroFileMessageDeserializer_DeserializeValueFromHeader(t *testing.T) {
	t.Parallel()

	schemaFile := writeTestSchemaFile(t)

	codec, err := goavro.NewCodecForStandardJSONFull(testAvroFileSchema)
	if err != nil {
		t.Fatalf("failed to create codec: %v", err)
	}

	native, _, err := codec.NativeFromTextual([]byte(`{"id":"abc"}`))
	if err != nil {
		t.Fatalf("failed to convert to native: %v", err)
	}

	binaryValue, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		t.Fatalf("failed to convert to binary: %v", err)
	}

	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	msg := &sarama.ConsumerMessage{
		Value: binaryValue,
		Headers: []*sarama.RecordHeader{
			{Key: []byte("ce_dataschema"), Value: []byte(fileURI(schemaFile))},
		},
	}

	data, err := deserializer.DeserializeValue(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `{"id":"abc"}`
	if string(data.data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data.data))
	}
}

func TestAvroFileMessageDeserializer_DeserializeValueFromEmptyHeader(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	msg := &sarama.ConsumerMessage{
		Value: []byte("some-avro-bytes"),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("ce_dataschema"), Value: []byte{}},
		},
	}

	_, err := deserializer.DeserializeValue(msg)
	if err == nil {
		t.Fatal("expected error for empty schema header value, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected error to mention the header being empty, got: %v", err)
	}
}

func TestAvroFileMessageDeserializer_DeserializeValueFromHeaderWithPlainPath(t *testing.T) {
	t.Parallel()

	schemaFile := writeTestSchemaFile(t)

	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	msg := &sarama.ConsumerMessage{
		Value: []byte("some-avro-bytes"),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("ce_dataschema"), Value: []byte(schemaFile)},
		},
	}

	_, err := deserializer.DeserializeValue(msg)
	if err == nil {
		t.Fatal("expected error for a plain path in the schema header, got nil")
	}
	if !isDeserializationError(err) {
		t.Fatalf("expected a DeserializationError, got: %v", err)
	}
}

func TestAvroFileMessageDeserializer_DeserializeValueInvalidPayload(t *testing.T) {
	t.Parallel()

	schemaFile := writeTestSchemaFile(t)

	deserializer := createTestDeserializer(t, schemaFile, "")

	_, err := deserializer.DeserializeValue(&sarama.ConsumerMessage{Value: []byte{0xff, 0xff, 0xff}})
	if err == nil {
		t.Fatal("expected error for invalid avro payload, got nil")
	}
	if !isDeserializationError(err) {
		t.Fatalf("expected a DeserializationError so that --ignore-errors can skip it, got: %v", err)
	}
}

func fileURI(path string) string {
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
}
