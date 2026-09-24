package consume

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestProtobufMessageConfiguredMarshalOptions(t *testing.T) {
	protoDir := t.TempDir()
	const protoFile = "configured-message.proto"
	const schema = `syntax = "proto3";

message ConfiguredMessage {
  enum Color {
    COLOR_UNSPECIFIED = 0;
    BLUE = 1;
  }
  message Details {
    string description = 1;
  }
  string display_name = 1;
  Color favorite_color = 2;
  int32 item_count = 3;
  Details details = 4;
}
`
	if err := os.WriteFile(filepath.Join(protoDir, protoFile), []byte(schema), 0o600); err != nil {
		t.Fatalf("failed to write protobuf fixture: %v", err)
	}

	const allOptions = `
      marshalOptions:
        allowPartial: true
        useProtoNames: true
        useEnumNumbers: true
        emitUnpopulated: true
        emitDefaultValues: true`

	for _, test := range []struct {
		name    string
		options string
		flags   []string
		want    string
	}{
		{
			name: "defaults",
			want: `{"displayName":"hello world","favoriteColor":"BLUE"}`,
		},
		{
			name: "configured_names_and_enum_numbers",
			options: `
      marshalOptions:
        useProtoNames: true
        useEnumNumbers: true`,
			want: `{"display_name":"hello world","favorite_color":1}`,
		},
		{
			name: "configured_default_values",
			options: `
      marshalOptions:
        emitDefaultValues: true`,
			want: `{"displayName":"hello world","favoriteColor":"BLUE","itemCount":0}`,
		},
		{
			name: "configured_unpopulated_fields",
			options: `
      marshalOptions:
        emitUnpopulated: true`,
			want: `{"displayName":"hello world","favoriteColor":"BLUE","itemCount":0,"details":null}`,
		},
		{
			name:    "all_configured_options",
			options: allOptions,
			want:    `{"display_name":"hello world","favorite_color":1,"item_count":0,"details":null}`,
		},
		{
			name:    "explicit_false_preserves_other_configured_options",
			options: allOptions,
			flags:   []string{"useProtoNames=false", "emitUnpopulated=false"},
			want:    `{"displayName":"hello world","favoriteColor":1,"itemCount":0}`,
		},
		{
			name:    "explicit_false_overrides_all_configured_options",
			options: allOptions,
			flags: []string{
				"allowPartial=false",
				"useProtoNames=false",
				"useEnumNumbers=false",
				"emitUnpopulated=false",
				"emitDefaultValues=false",
			},
			want: `{"displayName":"hello world","favoriteColor":"BLUE"}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			configYAML := fmt.Sprintf(`contexts:
  configured:
    brokers:
      - localhost:9092
    protobuf:
      importPaths:
        - %q
      protoFiles:
        - %q%s
current-context: configured
`, protoDir, protoFile, test.options)
			config := loadProtobufConfig(t, "configured", configYAML, nil)
			config, err := addFlagsToProtobufConfig(config, Flags{ProtoMarshalOptions: test.flags})
			if err != nil {
				t.Fatalf("failed to apply protobuf flags: %v", err)
			}

			descriptor := protobuf.ResolveMessageType(config, "ConfiguredMessage")
			if descriptor == nil {
				t.Fatal("failed to resolve configured protobuf message")
			}
			message := dynamicpb.NewMessage(descriptor)
			message.Set(descriptor.Fields().ByName("display_name"), protoreflect.ValueOfString("hello world"))
			message.Set(descriptor.Fields().ByName("favorite_color"), protoreflect.ValueOfEnum(1))
			wire, err := proto.Marshal(message)
			if err != nil {
				t.Fatalf("failed to encode protobuf message: %v", err)
			}

			deserializer, err := CreateProtobufMessageDeserializer(config, "ConfiguredMessage", "ConfiguredMessage")
			if err != nil {
				t.Fatalf("failed to create protobuf deserializer: %v", err)
			}
			consumerMessage := &sarama.ConsumerMessage{Key: wire, Value: wire}
			for _, path := range []struct {
				name        string
				deserialize func(*sarama.ConsumerMessage) (*DeserializedData, error)
			}{
				{name: "key", deserialize: deserializer.DeserializeKey},
				{name: "value", deserialize: deserializer.DeserializeValue},
			} {
				t.Run(path.name, func(t *testing.T) {
					result, err := path.deserialize(consumerMessage)
					if err != nil {
						t.Fatalf("failed to deserialize protobuf message: %v", err)
					}
					var got, want any
					if err := json.Unmarshal(result.data, &got); err != nil {
						t.Fatalf("failed to parse deserialized JSON %q: %v", result.data, err)
					}
					if err := json.Unmarshal([]byte(test.want), &want); err != nil {
						t.Fatalf("failed to parse expected JSON: %v", err)
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("unexpected protobuf JSON: got %s, want %s", result.data, test.want)
					}
				})
			}
		})
	}
}
