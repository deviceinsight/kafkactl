package consume

import (
	"encoding/base64"
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

// a schema uri taken from the message can be unreachable without anything being wrong with the
// setup, so --ignore-errors must be able to skip it rather than the whole consume aborting
func TestAvroFileMessageDeserializer_DeserializeValueFromHeaderWithUnresolvableURI(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	msg := &sarama.ConsumerMessage{
		Value: []byte("some-avro-bytes"),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("ce_dataschema"), Value: []byte(fileURI(filepath.Join(t.TempDir(), "gone.avsc")))},
		},
	}

	_, err := deserializer.DeserializeValue(msg)
	if err == nil {
		t.Fatal("expected error for an unresolvable schema uri, got nil")
	}
	if !isDeserializationError(err) {
		t.Fatalf("expected a DeserializationError so that --ignore-errors can skip it, got: %v", err)
	}
	if !strings.Contains(err.Error(), "ce_dataschema") {
		t.Fatalf("expected the error to name the header it came from, got: %v", err)
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

func TestAvroFileMessageDeserializer_CanDeserializeValueSkipsUnsupportedStructuredFormats(t *testing.T) {
	t.Parallel()

	schemaFile := writeTestSchemaFile(t)

	for _, contentType := range []string{
		"application/cloudevents-batch+json",
		"application/cloudevents+protobuf",
		"application/cloudevents+avro",
	} {
		for name, deserializer := range map[string]*AvroFileMessageDeserializer{
			"schema file":   createTestDeserializer(t, schemaFile, ""),
			"schema header": createTestDeserializer(t, "", "ce_dataschema"),
		} {
			msg := &sarama.ConsumerMessage{
				Value: []byte("some-event"),
				Headers: []*sarama.RecordHeader{
					{Key: []byte("content-type"), Value: []byte(contentType)},
					{Key: []byte("ce_dataschema"), Value: []byte(fileURI(schemaFile))},
				},
			}

			if deserializer.CanDeserializeValue(msg, Flags{}) {
				t.Errorf("%s: expected %s message to be skipped", name, contentType)
			}
		}
	}
}

func TestAvroFileMessageDeserializer_CanDeserializeValueWithStructuredJSON(t *testing.T) {
	t.Parallel()

	// no ce_dataschema header: in structured mode the schema is referenced from the event itself
	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	msg := &sarama.ConsumerMessage{
		Value: []byte(`{}`),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("content-type"), Value: []byte("Application/CloudEvents+JSON; charset=UTF-8")},
		},
	}

	if !deserializer.CanDeserializeValue(msg, Flags{}) {
		t.Error("expected structured mode json cloudevent to be deserialized")
	}
}

func TestAvroFileMessageDeserializer_DeserializeStructuredJSONFromDataschema(t *testing.T) {
	t.Parallel()

	schemaFile := writeTestSchemaFile(t)
	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	event := `{"specversion":"1.0","id":"1","source":"/test","type":"test","bignum":12345678901234567890,` +
		`"dataschema":"` + fileURI(schemaFile) + `","data_base64":"` + encodeTestAvroBase64(t, `{"id":"abc"}`) + `","ext":"x"}`

	data, err := deserializer.DeserializeValue(structuredJSONMessage(event))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `{"specversion":"1.0","id":"1","source":"/test","type":"test","bignum":12345678901234567890,` +
		`"dataschema":"` + fileURI(schemaFile) + `","data":{"id":"abc"},"ext":"x"}`
	if string(data.data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data.data))
	}
	if data.schema != testAvroFileSchema {
		t.Fatalf("expected schema to be set, got %q", data.schema)
	}
}

func TestAvroFileMessageDeserializer_DeserializeStructuredJSONWithSchemaFile(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, writeTestSchemaFile(t), "")

	event := `{"specversion":"1.0","data_base64":"` + encodeTestAvroBase64(t, `{"id":"abc"}`) + `"}`

	data, err := deserializer.DeserializeValue(structuredJSONMessage(event))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `{"specversion":"1.0","data":{"id":"abc"}}`
	if string(data.data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data.data))
	}
}

func TestAvroFileMessageDeserializer_DeserializeStructuredJSONFromHeader(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	msg := structuredJSONMessage(`{"specversion":"1.0","data_base64":"` + encodeTestAvroBase64(t, `{"id":"abc"}`) + `"}`)
	msg.Headers = append(msg.Headers, &sarama.RecordHeader{Key: []byte("ce_dataschema"), Value: []byte(fileURI(writeTestSchemaFile(t)))})

	data, err := deserializer.DeserializeValue(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `{"specversion":"1.0","data":{"id":"abc"}}`
	if string(data.data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data.data))
	}
}

func TestAvroFileMessageDeserializer_DeserializeStructuredJSONUnchanged(t *testing.T) {
	t.Parallel()

	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	for name, event := range map[string]string{
		"json data":                  `{"specversion":"1.0", "data": {"id": "abc"}}`,
		"data_base64 without schema": `{"specversion":"1.0", "data_base64": "AAAA"}`,
		"no data":                    `{"specversion":"1.0"}`,
	} {
		data, err := deserializer.DeserializeValue(structuredJSONMessage(event))
		if err != nil {
			t.Fatalf("%s: expected no error, got %v", name, err)
		}
		if string(data.data) != event {
			t.Errorf("%s: expected event to be unchanged, got %s", name, string(data.data))
		}
	}
}

func TestAvroFileMessageDeserializer_DeserializeStructuredJSONInvalid(t *testing.T) {
	t.Parallel()

	schemaFile := writeTestSchemaFile(t)
	missingSchema := fileURI(filepath.Join(t.TempDir(), "gone.avsc"))
	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	for name, event := range map[string]string{
		"not json":                 `not-json`,
		"not an object":            `["specversion"]`,
		"data and data_base64":     `{"data":{},"data_base64":"AAAA","dataschema":"` + fileURI(schemaFile) + `"}`,
		"invalid base64":           `{"data_base64":"%%%","dataschema":"` + fileURI(schemaFile) + `"}`,
		"invalid avro":             `{"data_base64":"//8=","dataschema":"` + fileURI(schemaFile) + `"}`,
		"dataschema plain path":    `{"data_base64":"AAAA","dataschema":"` + schemaFile + `"}`,
		"dataschema not a string":  `{"data_base64":"AAAA","dataschema":42}`,
		"data_base64 not a string": `{"data_base64":42,"dataschema":"` + fileURI(schemaFile) + `"}`,
		"unresolvable dataschema":  `{"data_base64":"AAAA","dataschema":"` + missingSchema + `"}`,
	} {
		_, err := deserializer.DeserializeValue(structuredJSONMessage(event))
		if err == nil {
			t.Errorf("%s: expected error, got nil", name)
			continue
		}
		if !isDeserializationError(err) {
			t.Errorf("%s: expected a DeserializationError, got: %v", name, err)
		}
	}
}

func structuredJSONMessage(event string) *sarama.ConsumerMessage {
	return &sarama.ConsumerMessage{
		Value: []byte(event),
		Headers: []*sarama.RecordHeader{
			{Key: []byte("content-type"), Value: []byte("application/cloudevents+json; charset=UTF-8")},
		},
	}
}

func encodeTestAvroBase64(t *testing.T, value string) string {
	t.Helper()

	codec, err := goavro.NewCodecForStandardJSONFull(testAvroFileSchema)
	if err != nil {
		t.Fatalf("failed to create codec: %v", err)
	}

	native, _, err := codec.NativeFromTextual([]byte(value))
	if err != nil {
		t.Fatalf("failed to convert to native: %v", err)
	}

	binaryValue, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		t.Fatalf("failed to convert to binary: %v", err)
	}

	return base64.StdEncoding.EncodeToString(binaryValue)
}

func TestAvroFileMessageDeserializer_CanDeserializeValueWithBinaryModeContentType(t *testing.T) {
	t.Parallel()

	schemaFile := writeTestSchemaFile(t)

	deserializer := createTestDeserializer(t, "", "ce_dataschema")

	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("content-type"), Value: []byte("application/avro")},
			{Key: []byte("ce_dataschema"), Value: []byte(fileURI(schemaFile))},
		},
	}

	if !deserializer.CanDeserializeValue(msg, Flags{}) {
		t.Error("expected binary mode cloudevent to be deserialized")
	}
}
