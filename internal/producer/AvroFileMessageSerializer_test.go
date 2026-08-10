package producer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
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

func TestAvroFileMessageSerializer_CanSerializeKeyAlwaysFalse(t *testing.T) {
	t.Parallel()

	serializer, err := CreateAvroFileMessageSerializer(writeTestSchemaFile(t), avro.NewSchemaSource(avro.Standard))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	canSerialize, err := serializer.CanSerializeKey("topic")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if canSerialize {
		t.Fatal("expected CanSerializeKey to always return false")
	}
}

func TestAvroFileMessageSerializer_CanSerializeValue(t *testing.T) {
	t.Parallel()

	serializer, err := CreateAvroFileMessageSerializer(writeTestSchemaFile(t), avro.NewSchemaSource(avro.Standard))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	canSerialize, err := serializer.CanSerializeValue("topic")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !canSerialize {
		t.Fatal("expected CanSerializeValue to return true when a schema was resolved")
	}
}

func TestAvroFileMessageSerializer_SerializeValueRoundTrip(t *testing.T) {
	t.Parallel()

	serializer, err := CreateAvroFileMessageSerializer(writeTestSchemaFile(t), avro.NewSchemaSource(avro.Standard))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	encoded, err := serializer.SerializeValue([]byte(`{"id":"abc"}`), Flags{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	native, _, err := serializer.codec.NativeFromBinary(encoded)
	if err != nil {
		t.Fatalf("failed to decode encoded value: %v", err)
	}

	textual, err := serializer.codec.TextualFromNative(nil, native)
	if err != nil {
		t.Fatalf("failed to convert decoded value to textual: %v", err)
	}

	expected := `{"id":"abc"}`
	if string(textual) != expected {
		t.Fatalf("expected %s, got %s", expected, string(textual))
	}
}

func TestAvroFileMessageSerializer_CreateFailsForMissingSchema(t *testing.T) {
	t.Parallel()

	if _, err := CreateAvroFileMessageSerializer(filepath.Join(t.TempDir(), "missing.avsc"), avro.NewSchemaSource(avro.Standard)); err == nil {
		t.Fatal("expected error for missing schema file, got nil")
	}
}
