package consume

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
)

const rawAvroTestSchema = `{
  "type": "record",
  "name": "person",
  "fields": [
    {"name": "name", "type": "string"}
  ]
}`

func TestRawAvroMessageDeserializerUsesStaticSchema(t *testing.T) {
	const wantData = `{"name":"Alice"}`

	codec := newRawAvroTestCodec(t, rawAvroTestSchema)
	payload := encodeRawAvroTestValue(t, codec, wantData)
	deserializer := NewRawAvroMessageDeserializer("", codec, nil)
	msg := &sarama.ConsumerMessage{Key: []byte("plain-key"), Value: payload}

	if deserializer.CanDeserializeKey(msg, Flags{}) {
		t.Fatal("CanDeserializeKey() = true, want false")
	}
	if !deserializer.CanDeserializeValue(msg, Flags{}) {
		t.Fatal("CanDeserializeValue() = false for a non-nil value, want true")
	}

	decoded, err := deserializer.DeserializeValue(msg)
	if err != nil {
		t.Fatalf("DeserializeValue(): unexpected error: %v", err)
	}
	if string(decoded.data) != wantData {
		t.Fatalf("DeserializeValue() data = %q, want %q", decoded.data, wantData)
	}
	if decoded.schema != rawAvroTestSchema {
		t.Fatalf("DeserializeValue() schema = %q, want %q", decoded.schema, rawAvroTestSchema)
	}
	if decoded.schemaID != nil {
		t.Fatalf("DeserializeValue() schemaID = %d, want nil", *decoded.schemaID)
	}
}

func TestRawAvroMessageDeserializerUsesLastMatchingHeader(t *testing.T) {
	codec := newRawAvroTestCodec(t, rawAvroTestSchema)
	payload := encodeRawAvroTestValue(t, codec, `{"name":"Bob"}`)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(writer, rawAvroTestSchema)
	}))
	defer server.Close()

	resolver := avro.NewSchemaResolver(avro.Standard, true)
	deserializer := NewRawAvroMessageDeserializer("ce_dataschema", nil, resolver)
	msg := &sarama.ConsumerMessage{
		Value: payload,
		Headers: []*sarama.RecordHeader{
			{Key: []byte("ce_dataschema"), Value: []byte("http://127.0.0.1:1/unused")},
			{Key: []byte("other"), Value: []byte("ignored")},
			{Key: []byte("ce_dataschema"), Value: []byte(server.URL)},
		},
	}

	decoded, err := deserializer.DeserializeValue(msg)
	if err != nil {
		t.Fatalf("DeserializeValue(): unexpected error: %v", err)
	}
	const wantData = `{"name":"Bob"}`
	if string(decoded.data) != wantData {
		t.Fatalf("DeserializeValue() data = %q, want %q", decoded.data, wantData)
	}
}

func TestRawAvroMessageDeserializerDecodesAMQPHeaderValue(t *testing.T) {
	codec := newRawAvroTestCodec(t, rawAvroTestSchema)
	payload := encodeRawAvroTestValue(t, codec, `{"name":"Carol"}`)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(writer, rawAvroTestSchema)
	}))
	defer server.Close()

	wrappedURL := append([]byte{0xa1, byte(len(server.URL))}, []byte(server.URL)...)
	deserializer := NewRawAvroMessageDeserializer(
		"ce_dataschema",
		nil,
		avro.NewSchemaResolver(avro.Standard, true),
	)
	msg := &sarama.ConsumerMessage{
		Value:   payload,
		Headers: []*sarama.RecordHeader{{Key: []byte("ce_dataschema"), Value: wrappedURL}},
	}

	decoded, err := deserializer.DeserializeValue(msg)
	if err != nil {
		t.Fatalf("DeserializeValue(): unexpected error: %v", err)
	}
	const wantData = `{"name":"Carol"}`
	if string(decoded.data) != wantData {
		t.Fatalf("DeserializeValue() data = %q, want %q", decoded.data, wantData)
	}
}

func TestRawAvroMessageDeserializerFallsBackToStaticSchemaWhenHeaderIsMissing(t *testing.T) {
	codec := newRawAvroTestCodec(t, rawAvroTestSchema)
	payload := encodeRawAvroTestValue(t, codec, `{"name":"Dora"}`)
	deserializer := NewRawAvroMessageDeserializer(
		"ce_dataschema",
		codec,
		avro.NewSchemaResolver(avro.Standard, true),
	)

	decoded, err := deserializer.DeserializeValue(&sarama.ConsumerMessage{Value: payload})
	if err != nil {
		t.Fatalf("DeserializeValue(): unexpected error: %v", err)
	}
	const wantData = `{"name":"Dora"}`
	if string(decoded.data) != wantData {
		t.Fatalf("DeserializeValue() data = %q, want %q", decoded.data, wantData)
	}
}

func TestRawAvroMessageDeserializerRejectsInvalidHeader(t *testing.T) {
	codec := newRawAvroTestCodec(t, rawAvroTestSchema)
	payload := encodeRawAvroTestValue(t, codec, `{"name":"Eve"}`)
	deserializer := NewRawAvroMessageDeserializer(
		"ce_dataschema",
		codec,
		avro.NewSchemaResolver(avro.Standard, true),
	)
	msg := &sarama.ConsumerMessage{
		Value:   payload,
		Headers: []*sarama.RecordHeader{{Key: []byte("ce_dataschema"), Value: []byte("file:///tmp/schema.avsc")}},
	}

	_, err := deserializer.DeserializeValue(msg)
	const wantError = "unsupported avro schema URI scheme"
	if err == nil {
		t.Fatalf("DeserializeValue() error = nil, want it to contain %q", wantError)
	}
	if !strings.Contains(err.Error(), wantError) {
		t.Fatalf("DeserializeValue() error = %q, want it to contain %q", err, wantError)
	}
}

func TestRawAvroMessageDeserializerRejectsMissingOrEmptyHeader(t *testing.T) {
	deserializer := NewRawAvroMessageDeserializer(
		"ce_dataschema",
		nil,
		avro.NewSchemaResolver(avro.Standard, true),
	)

	_, err := deserializer.DeserializeValue(&sarama.ConsumerMessage{Value: []byte{1}})
	const wantMissingError = `avro schema header "ce_dataschema" is missing`
	if err == nil {
		t.Fatalf("DeserializeValue() error = nil, want it to contain %q", wantMissingError)
	}
	if !strings.Contains(err.Error(), wantMissingError) {
		t.Fatalf("DeserializeValue() error = %q, want it to contain %q", err, wantMissingError)
	}

	_, err = deserializer.DeserializeValue(&sarama.ConsumerMessage{
		Value:   []byte{1},
		Headers: []*sarama.RecordHeader{{Key: []byte("ce_dataschema"), Value: []byte("  ")}},
	})
	const wantEmptyError = `avro schema header "ce_dataschema" is empty`
	if err == nil {
		t.Fatalf("DeserializeValue() error = nil, want it to contain %q", wantEmptyError)
	}
	if !strings.Contains(err.Error(), wantEmptyError) {
		t.Fatalf("DeserializeValue() error = %q, want it to contain %q", err, wantEmptyError)
	}
}

func TestRawAvroMessageDeserializerHeaderNameIsCaseSensitive(t *testing.T) {
	deserializer := NewRawAvroMessageDeserializer(
		"ce_dataschema",
		nil,
		avro.NewSchemaResolver(avro.Standard, true),
	)
	msg := &sarama.ConsumerMessage{
		Value:   []byte{1},
		Headers: []*sarama.RecordHeader{{Key: []byte("CE_DATASCHEMA"), Value: []byte("http://example.com")}},
	}

	_, err := deserializer.DeserializeValue(msg)
	const wantError = `avro schema header "ce_dataschema" is missing`
	if err == nil {
		t.Fatalf("DeserializeValue() error = nil, want it to contain %q", wantError)
	}
	if !strings.Contains(err.Error(), wantError) {
		t.Fatalf("DeserializeValue() error = %q, want it to contain %q", err, wantError)
	}
}

func TestRawAvroMessageDeserializerDoesNotHandleTombstone(t *testing.T) {
	deserializer := NewRawAvroMessageDeserializer(
		"ce_dataschema",
		nil,
		avro.NewSchemaResolver(avro.Standard, true),
	)
	msg := &sarama.ConsumerMessage{Value: nil}

	if deserializer.CanDeserializeValue(msg, Flags{}) {
		t.Fatal("CanDeserializeValue() = true for a nil tombstone, want false")
	}
}

func TestRawAvroMessageDeserializerDecodesNonNilZeroByteValue(t *testing.T) {
	const (
		schema   = `"null"`
		wantData = "null"
	)

	codec := newRawAvroTestCodec(t, schema)
	payload := encodeRawAvroTestValue(t, codec, wantData)
	if payload == nil {
		t.Fatal("EncodeTextual() payload = nil, want a non-nil zero-byte value")
	}
	if len(payload) != 0 {
		t.Fatalf("EncodeTextual() payload length = %d, want 0", len(payload))
	}

	deserializer := NewRawAvroMessageDeserializer("", codec, nil)
	msg := &sarama.ConsumerMessage{Value: payload}
	if !deserializer.CanDeserializeValue(msg, Flags{}) {
		t.Fatal("CanDeserializeValue() = false for a non-nil zero-byte value, want true")
	}

	decoded, err := deserializer.DeserializeValue(msg)
	if err != nil {
		t.Fatalf("DeserializeValue(): unexpected error: %v", err)
	}
	if string(decoded.data) != wantData {
		t.Fatalf("DeserializeValue() data = %q, want %q", decoded.data, wantData)
	}
	if decoded.schema != schema {
		t.Fatalf("DeserializeValue() schema = %q, want %q", decoded.schema, schema)
	}
}

func TestRawAvroMessageDeserializerPropagatesDecodeError(t *testing.T) {
	codec := newRawAvroTestCodec(t, rawAvroTestSchema)
	deserializer := NewRawAvroMessageDeserializer("", codec, nil)
	msg := &sarama.ConsumerMessage{Value: []byte{2}}

	decoded, err := deserializer.DeserializeValue(msg)
	if err == nil {
		t.Fatalf("expected DeserializeValue() to return a decode error, got nil with data %#v", decoded)
	}
	const wantError = "failed to decode avro data"
	if !strings.Contains(err.Error(), wantError) {
		t.Fatalf("DeserializeValue() error = %q, want it to contain %q", err, wantError)
	}
}

func newRawAvroTestCodec(t *testing.T, schema string) *avro.MessageCodec {
	t.Helper()

	codec, err := avro.NewMessageCodec(schema, avro.Standard)
	if err != nil {
		t.Fatalf("NewMessageCodec(): unexpected error: %v", err)
	}
	return codec
}

func encodeRawAvroTestValue(t *testing.T, codec *avro.MessageCodec, value string) []byte {
	t.Helper()

	payload, err := codec.EncodeTextual([]byte(value))
	if err != nil {
		t.Fatalf("EncodeTextual(): unexpected error: %v", err)
	}
	return payload
}
