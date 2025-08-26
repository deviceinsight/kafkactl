package producer

import (
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type ProtobufMessageSerializer struct {
	topic           string
	keyDescriptor   protoreflect.MessageDescriptor
	valueDescriptor protoreflect.MessageDescriptor
}

func CreateProtobufMessageSerializer(topic string, context internal.ProtobufConfig, keyType, valueType protoreflect.FullName) (*ProtobufMessageSerializer, error) {
	valueDescriptor := protobuf.ResolveMessageType(context, valueType)
	if valueDescriptor == nil && valueType != "" {
		return nil, errors.Errorf("value message type %q not found in provided files", valueType)
	}

	keyDescriptor := protobuf.ResolveMessageType(context, keyType)
	if keyDescriptor == nil && keyType != "" {
		return nil, errors.Errorf("key message type %q not found in provided files", keyType)
	}

	return &ProtobufMessageSerializer{
		topic:           topic,
		keyDescriptor:   keyDescriptor,
		valueDescriptor: valueDescriptor,
	}, nil
}

func (serializer ProtobufMessageSerializer) CanSerializeValue(_ string) (bool, error) {
	return serializer.valueDescriptor != nil, nil
}

func (serializer ProtobufMessageSerializer) CanSerializeKey(_ string) (bool, error) {
	return serializer.keyDescriptor != nil, nil
}

func (serializer ProtobufMessageSerializer) SerializeValue(value []byte, _ Flags) ([]byte, error) {
	return encodeProtobuf(value, serializer.valueDescriptor)
}

func (serializer ProtobufMessageSerializer) SerializeKey(key []byte, _ Flags) ([]byte, error) {
	return encodeProtobuf(key, serializer.keyDescriptor)
}

func encodeProtobuf(data []byte, messageDescriptor protoreflect.MessageDescriptor) ([]byte, error) {
	if messageDescriptor == nil {
		return data, nil
	}

	message := dynamicpb.NewMessage(messageDescriptor)

	unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}

	if err := unmarshaler.Unmarshal(data, message); err != nil {
		return nil, errors.Wrap(err, "invalid json")
	}

	pb, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}

	return pb, nil
}
