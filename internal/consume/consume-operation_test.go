package consume

import (
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal"
)

func TestValidateRawAvroFlags(t *testing.T) {
	tests := []struct {
		name           string
		flags          Flags
		protobufConfig internal.ProtobufConfig
		wantError      string
	}{
		{
			name: "schema file with value protobuf",
			flags: Flags{
				AvroSchemaFile: "event.avsc",
				ValueProtoType: "example.Event",
			},
			wantError: "parameter --value-proto-type cannot be used with --avro-schema-file or --avro-schema-header",
		},
		{
			name: "schema header with value protobuf",
			flags: Flags{
				AvroSchemaHeader: "ce_dataschema",
				ValueProtoType:   "example.Event",
			},
			wantError: "parameter --value-proto-type cannot be used with --avro-schema-file or --avro-schema-header",
		},
		{
			name: "schema file and header with value protobuf",
			flags: Flags{
				AvroSchemaFile:   "event.avsc",
				AvroSchemaHeader: "ce_dataschema",
				ValueProtoType:   "example.Event",
			},
			wantError: "parameter --value-proto-type cannot be used with --avro-schema-file or --avro-schema-header",
		},
		{
			name: "key protobuf with proto file",
			flags: Flags{
				AvroSchemaFile: "event.avsc",
				KeyProtoType:   "example.Key",
			},
			protobufConfig: internal.ProtobufConfig{ProtoFiles: []string{"event.proto"}},
		},
		{
			name: "key protobuf with protoset file",
			flags: Flags{
				AvroSchemaHeader: "ce_dataschema",
				KeyProtoType:     "example.Key",
			},
			protobufConfig: internal.ProtobufConfig{ProtosetFiles: []string{"event.protoset"}},
		},
		{
			name: "key protobuf without local description",
			flags: Flags{
				AvroSchemaFile: "event.avsc",
				KeyProtoType:   "example.Key",
			},
			wantError: "parameter --key-proto-type requires a local protobuf description file " +
				"when used with --avro-schema-file or --avro-schema-header",
		},
		{
			name: "key protobuf with import path only",
			flags: Flags{
				AvroSchemaHeader: "ce_dataschema",
				KeyProtoType:     "example.Key",
			},
			protobufConfig: internal.ProtobufConfig{ProtoImportPaths: []string{"proto"}},
			wantError: "parameter --key-proto-type requires a local protobuf description file " +
				"when used with --avro-schema-file or --avro-schema-header",
		},
		{
			name:  "non-raw key protobuf without local description",
			flags: Flags{KeyProtoType: "example.Key"},
		},
		{
			name:  "non-raw value protobuf without local description",
			flags: Flags{ValueProtoType: "example.Event"},
		},
		{
			name: "schema cache alone with value protobuf",
			flags: Flags{
				AvroSchemaCache: true,
				ValueProtoType:  "example.Event",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRawAvroFlags(test.flags, test.protobufConfig)
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("validateRawAvroFlags(): unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected validateRawAvroFlags() to return error %q, got nil", test.wantError)
			}
			if err.Error() != test.wantError {
				t.Fatalf("validateRawAvroFlags() error = %q, want %q", err, test.wantError)
			}
		})
	}
}
