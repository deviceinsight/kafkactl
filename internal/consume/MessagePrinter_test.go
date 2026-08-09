package consume

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"gopkg.in/yaml.v2"
)

func TestPrintMessageSchemaMetadata(t *testing.T) {
	const (
		schema = `{"type":"string"}`
		value  = `"hello"`
	)
	schemaID := 42

	for _, test := range []struct {
		name     string
		schema   string
		schemaID *int
		want     string
	}{
		{
			name:   "raw Avro",
			schema: schema,
			want:   `||{"type":"string"}||"hello"` + "\n",
		},
		{
			name:     "Schema Registry",
			schema:   schema,
			schemaID: &schemaID,
			want:     `||{"type":"string"}|42|"hello"` + "\n",
		},
		{
			name:     "schema ID without schema",
			schemaID: &schemaID,
			want:     `|||42|"hello"` + "\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			flags := Flags{PrintSchema: true, Separator: "|"}
			msg := newMessage(
				&sarama.ConsumerMessage{},
				flags,
				nil,
				&DeserializedData{schema: test.schema, schemaID: test.schemaID, data: []byte(value)},
			)

			got := printMessageOutput(t, msg, flags)
			if got != test.want {
				t.Fatalf("expected output %q, got %q", test.want, got)
			}
		})
	}
}

func TestPrintMessageRawAvroStructuredOutput(t *testing.T) {
	const (
		schema = `{"type":"string"}`
		value  = `"hello"`
	)

	for _, test := range []struct {
		name      string
		format    string
		unmarshal func([]byte, any) error
	}{
		{name: "JSON", format: "json", unmarshal: json.Unmarshal},
		{name: "YAML", format: "yaml", unmarshal: yaml.Unmarshal},
	} {
		t.Run(test.name, func(t *testing.T) {
			flags := Flags{PrintSchema: true, OutputFormat: test.format}
			msg := newMessage(
				&sarama.ConsumerMessage{},
				flags,
				nil,
				&DeserializedData{schema: schema, data: []byte(value)},
			)

			printed := printMessageOutput(t, msg, flags)
			fields := make(map[string]any)
			if err := test.unmarshal([]byte(printed), &fields); err != nil {
				t.Fatalf("failed to decode %s output %q: %v", test.name, printed, err)
			}
			gotSchema, found := fields["valueSchema"]
			if !found {
				t.Fatalf("expected %s output to contain valueSchema, got %q", test.name, printed)
			}
			if gotSchema != schema {
				t.Fatalf("expected %s valueSchema %q, got %q", test.name, schema, gotSchema)
			}
			if gotSchemaID, found := fields["valueSchemaId"]; found {
				t.Fatalf("expected %s output to omit valueSchemaId, got %v", test.name, gotSchemaID)
			}
		})
	}
}

func printMessageOutput(t *testing.T, msg *message, flags Flags) string {
	t.Helper()

	previousStreams := output.IoStreams
	var stdout bytes.Buffer
	output.IoStreams.Out = &stdout
	defer func() {
		output.IoStreams = previousStreams
	}()

	if err := printMessage(msg, flags); err != nil {
		t.Fatalf("printMessage(): unexpected error: %v", err)
	}
	return stdout.String()
}
