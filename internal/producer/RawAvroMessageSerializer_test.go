package producer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/linkedin/goavro/v2"
)

const rawAvroProducerTestSchema = `{
  "type": "record",
  "name": "person",
  "fields": [
    {"name": "name", "type": "string"}
  ]
}`

func TestRawAvroMessageSerializerEncodesUnframedValue(t *testing.T) {
	messageCodec, err := avro.NewMessageCodec(rawAvroProducerTestSchema, avro.Standard)
	if err != nil {
		t.Fatalf("failed to create message codec: %v", err)
	}
	serializer := NewRawAvroMessageSerializer(messageCodec)
	value := []byte(`{"name":"Alice"}`)

	encoded, err := serializer.SerializeValue(value, Flags{})
	if err != nil {
		t.Fatalf("failed to serialize raw Avro value: %v", err)
	}

	expectedCodec, err := goavro.NewCodecForStandardJSONFull(rawAvroProducerTestSchema)
	if err != nil {
		t.Fatalf("failed to create expected codec: %v", err)
	}
	native, _, err := expectedCodec.NativeFromTextual(value)
	if err != nil {
		t.Fatalf("failed to create expected native value: %v", err)
	}
	expected, err := expectedCodec.BinaryFromNative(nil, native)
	if err != nil {
		t.Fatalf("failed to create expected binary value: %v", err)
	}

	if !bytes.Equal(encoded, expected) {
		t.Fatalf("raw Avro output differs from unframed goavro output: expected %v, got %v", expected, encoded)
	}
}

func TestRawAvroMessageSerializerOnlyHandlesValues(t *testing.T) {
	messageCodec, err := avro.NewMessageCodec(rawAvroProducerTestSchema, avro.Standard)
	if err != nil {
		t.Fatalf("failed to create message codec: %v", err)
	}
	serializer := NewRawAvroMessageSerializer(messageCodec)

	canSerializeValue, err := serializer.CanSerializeValue("topic")
	if err != nil || !canSerializeValue {
		t.Fatalf("expected serializer to handle values, canSerialize=%v err=%v", canSerializeValue, err)
	}
	canSerializeKey, err := serializer.CanSerializeKey("topic")
	if err != nil || canSerializeKey {
		t.Fatalf("expected serializer not to handle keys, canSerialize=%v err=%v", canSerializeKey, err)
	}
}

func TestRawAvroMessageSerializerPreservesTombstone(t *testing.T) {
	messageCodec, err := avro.NewMessageCodec(rawAvroProducerTestSchema, avro.Standard)
	if err != nil {
		t.Fatalf("failed to create message codec: %v", err)
	}
	serializer := NewRawAvroMessageSerializer(messageCodec)

	encoded, err := serializer.SerializeValue(nil, Flags{})
	if err != nil {
		t.Fatalf("failed to preserve tombstone: %v", err)
	}
	if encoded != nil {
		t.Fatalf("expected nil tombstone, got %v", encoded)
	}
}

func TestRawAvroMessageSerializerRejectsEmptyNonNilInput(t *testing.T) {
	messageCodec, err := avro.NewMessageCodec(rawAvroProducerTestSchema, avro.Standard)
	if err != nil {
		t.Fatalf("failed to create message codec: %v", err)
	}
	serializer := NewRawAvroMessageSerializer(messageCodec)

	_, err = serializer.SerializeValue([]byte{}, Flags{})
	if err == nil {
		t.Fatal("expected non-nil empty input to return an error")
	}
	const wantErrPart = "failed to convert textual data to avro data"
	if !strings.Contains(err.Error(), wantErrPart) {
		t.Fatalf("SerializeValue() error = %q, want it to contain %q", err, wantErrPart)
	}
}

func TestRawAvroMessageSerializerProducesNonNilZeroByteValue(t *testing.T) {
	messageCodec, err := avro.NewMessageCodec(
		`{"type":"record","name":"empty","fields":[]}`,
		avro.Standard,
	)
	if err != nil {
		t.Fatalf("failed to create message codec: %v", err)
	}
	serializer := NewRawAvroMessageSerializer(messageCodec)

	encoded, err := serializer.SerializeValue([]byte(`{}`), Flags{})
	if err != nil {
		t.Fatalf("failed to serialize zero-byte Avro value: %v", err)
	}
	if encoded == nil {
		t.Fatal("zero-byte Avro value must not be encoded as a nil Kafka tombstone")
	}
	if len(encoded) != 0 {
		t.Fatalf("expected zero encoded bytes, got %v", encoded)
	}
}
