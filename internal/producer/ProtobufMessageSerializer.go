package producer

import (
	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

type ProtobufMessageSerializer struct {
	topic           string
	keyDescriptor   *desc.MessageDescriptor
	valueDescriptor *desc.MessageDescriptor
}

func CreateProtobufMessageSerializer(topic string, context internal.ProtobufConfig, keyType, valueType string) (*ProtobufMessageSerializer, error) {
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

func (serializer ProtobufMessageSerializer) CanSerializeValue(topic string) (bool, error) {
	return serializer.valueDescriptor != nil, nil
}

func (serializer ProtobufMessageSerializer) CanSerializeKey(topic string) (bool, error) {
	return serializer.keyDescriptor != nil, nil
}

func (serializer ProtobufMessageSerializer) SerializeValue(value []byte, _ Flags) ([]byte, error) {
	return encodeProtobuf(value, serializer.valueDescriptor)
}

func (serializer ProtobufMessageSerializer) SerializeKey(key []byte, _ Flags) ([]byte, error) {
	return encodeProtobuf(key, serializer.keyDescriptor)
}

func encodeProtobuf(data []byte, messageDescriptor *desc.MessageDescriptor) (sarama.ByteEncoder, error) {
	if messageDescriptor == nil {
		return data, nil
	}

	message := dynamic.NewMessage(messageDescriptor)

	// can probably be replaced by:
	// umar := protojson.UnmarshalOptions{DiscardUnknown: true}

	if err := message.UnmarshalJSONPB(&jsonpb.Unmarshaler{AllowUnknownFields: true}, data); err != nil {
		return nil, errors.Wrap(err, "invalid json")
	}

	pb, err := message.Marshal()
	if err != nil {
		return nil, err
	}

	return pb, nil
}
