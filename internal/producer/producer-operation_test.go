package producer

import (
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal"
)

func TestValidateRawAvroFlagsAllowsKeyProtobufWithLocalDescription(t *testing.T) {
	for _, test := range []struct {
		name           string
		protobufConfig internal.ProtobufConfig
	}{
		{name: "proto file", protobufConfig: internal.ProtobufConfig{ProtoFiles: []string{"event.proto"}}},
		{name: "protoset file", protobufConfig: internal.ProtobufConfig{ProtosetFiles: []string{"event.protoset"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validateRawAvroFlags(Flags{
				AvroSchemaFile:     "event.avsc",
				KeyProtoType:       "example.Key",
				KeySchemaVersion:   -1,
				ValueSchemaVersion: -1,
			}, test.protobufConfig)
			if err != nil {
				t.Fatalf("validateRawAvroFlags(): unexpected error: %v", err)
			}
		})
	}
}

func TestValidateRawAvroFlagsRejectsKeyProtobufWithoutLocalDescription(t *testing.T) {
	for _, test := range []struct {
		name           string
		protobufConfig internal.ProtobufConfig
	}{
		{name: "no local description"},
		{name: "import path only", protobufConfig: internal.ProtobufConfig{ProtoImportPaths: []string{"proto"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validateRawAvroFlags(Flags{
				AvroSchemaFile:     "event.avsc",
				KeyProtoType:       "example.Key",
				KeySchemaVersion:   -1,
				ValueSchemaVersion: -1,
			}, test.protobufConfig)
			const wantErrPart = "--key-proto-type requires a local protobuf description file"
			if err == nil || !strings.Contains(err.Error(), wantErrPart) {
				t.Fatalf("validateRawAvroFlags() error = %q, want it to contain %q", err, wantErrPart)
			}
		})
	}
}

func TestValidateRawAvroFlagsAllowsRegistryKeyProtobufWithoutRawAvro(t *testing.T) {
	err := validateRawAvroFlags(Flags{KeyProtoType: "example.Key"}, internal.ProtobufConfig{})
	if err != nil {
		t.Fatalf("validateRawAvroFlags(): unexpected error: %v", err)
	}
}

func TestValidateRawAvroFlagsRejectsConflictingValueOptions(t *testing.T) {
	tests := []struct {
		name     string
		flags    Flags
		expected string
	}{
		{
			name: "value protobuf",
			flags: Flags{
				AvroSchemaFile:     "event.avsc",
				ValueProtoType:     "example.Event",
				KeySchemaVersion:   -1,
				ValueSchemaVersion: -1,
			},
			expected: "--value-proto-type",
		},
		{
			name: "value schema version",
			flags: Flags{
				AvroSchemaFile:     "event.avsc",
				KeySchemaVersion:   -1,
				ValueSchemaVersion: 2,
			},
			expected: "--value-schema-version",
		},
		{
			name: "key schema version",
			flags: Flags{
				AvroSchemaFile:     "event.avsc",
				KeySchemaVersion:   3,
				ValueSchemaVersion: -1,
			},
			expected: "--key-schema-version",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRawAvroFlags(test.flags, internal.ProtobufConfig{})
			if err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("expected conflict containing %q, got: %v", test.expected, err)
			}
		})
	}
}

func TestValidateProducerFlags(t *testing.T) {
	const separatorError = "separator is used to split input from stdin/file. it cannot be used together with key or value"

	for _, test := range []struct {
		name      string
		flags     Flags
		wantError string
	}{
		{name: "valid", flags: Flags{}},
		{name: "separator with key", flags: Flags{Separator: ",", Key: "key"}, wantError: separatorError},
		{name: "separator with value", flags: Flags{Separator: ",", Value: "value"}, wantError: separatorError},
		{
			name:      "null value with value",
			flags:     Flags{NullValue: true, Value: "value"},
			wantError: "parameters --null-value and --value cannot be used together",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := validateProducerFlags(test.flags, internal.ProtobufConfig{})
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("validateProducerFlags(): unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("validateProducerFlags() error = %q, want %q", err, test.wantError)
			}
		})
	}
}
