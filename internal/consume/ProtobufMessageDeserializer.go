package consume

import (
	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

type ProtobufMessageDeserializer struct {
	keyType         protoreflect.Name
	valueType       protoreflect.Name
	marshalOptions  internal.ProtobufMarshalOptions
	keyDescriptor   protoreflect.MessageDescriptor
	valueDescriptor protoreflect.MessageDescriptor
}

func CreateProtobufMessageDeserializer(config internal.ProtobufConfig, keyType, valueType protoreflect.FullName) (*ProtobufMessageDeserializer, error) {
	return &ProtobufMessageDeserializer{
		keyType:         keyType.Name(),
		valueType:       valueType.Name(),
		marshalOptions:  config.MarshalOptions,
		keyDescriptor:   protobuf.ResolveMessageType(config, keyType),
		valueDescriptor: protobuf.ResolveMessageType(config, valueType),
	}, nil
}

func (deserializer *ProtobufMessageDeserializer) CanDeserializeKey(_ *sarama.ConsumerMessage, flags Flags) bool {
	return flags.KeyProtoType != ""
}

func (deserializer *ProtobufMessageDeserializer) CanDeserializeValue(_ *sarama.ConsumerMessage, flags Flags) bool {
	return flags.ValueProtoType != ""
}

func (deserializer *ProtobufMessageDeserializer) DeserializeKey(consumerMsg *sarama.ConsumerMessage) (*DeserializedData, error) {
	output.Debugf("deserialize key with ProtobufMessageDeserializer")

	if deserializer.keyDescriptor == nil {
		return nil, errors.Errorf("key message type %q not found in provided files", deserializer.keyType)
	}

	deserialized, err := decodeProtobuf(consumerMsg.Key, deserializer.keyDescriptor, deserializer.marshalOptions)
	return &DeserializedData{data: deserialized}, err
}

func (deserializer *ProtobufMessageDeserializer) DeserializeValue(consumerMsg *sarama.ConsumerMessage) (*DeserializedData, error) {
	output.Debugf("deserialize value with ProtobufMessageDeserializer")
	if deserializer.valueDescriptor == nil {
		return nil, errors.Errorf("value message type %q not found in provided files", deserializer.valueType)
	}
	deserialized, err := decodeProtobuf(consumerMsg.Value, deserializer.valueDescriptor, deserializer.marshalOptions)
	return &DeserializedData{data: deserialized}, err
}

func decodeProtobuf(rawData []byte, messageDescriptor protoreflect.MessageDescriptor, marshalOptions internal.ProtobufMarshalOptions) ([]byte, error) {
	if len(rawData) == 0 { // tombstone record, can't be unmarshalled as protobuf
		return nil, nil
	}

	msg := dynamicpb.NewMessage(messageDescriptor)
	if err := proto.Unmarshal(rawData, msg); err != nil {
		return nil, err
	}

	jsonBytes, err := protojson.MarshalOptions{
		Multiline:         false,
		AllowPartial:      marshalOptions.AllowPartial,
		UseProtoNames:     marshalOptions.UseProtoNames,
		UseEnumNumbers:    marshalOptions.UseEnumNumbers,
		EmitUnpopulated:   marshalOptions.EmitUnpopulated,
		EmitDefaultValues: marshalOptions.EmitDefaultValues,
		Resolver:          &ignoreUnrecognizedAny{protoregistry.GlobalTypes},
	}.Marshal(msg)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal protbuf message")
	}

	return jsonBytes, nil
}
