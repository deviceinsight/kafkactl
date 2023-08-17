package producer

import (
	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/internal/helpers/protobuf"
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

func CreateProtobufMessageSerializer(topic string, context protobuf.SearchContext, keyType, valueType string) (*ProtobufMessageSerializer, error) {
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

func (serializer ProtobufMessageSerializer) CanSerialize(string) (bool, error) {
	return true, nil
}

func (serializer ProtobufMessageSerializer) Serialize(key, value []byte, flags Flags) (*sarama.ProducerMessage, error) {
	recordHeaders, err := createRecordHeaders(flags)
	if err != nil {
		return nil, err
	}

	message := &sarama.ProducerMessage{Topic: serializer.topic, Partition: flags.Partition, Headers: recordHeaders}

	if key != nil {
		message.Key, err = encodeProtobuf(key, serializer.keyDescriptor, flags.KeyEncoding)
		if err != nil {
			return nil, err
		}
	}

	message.Value, err = encodeProtobuf(value, serializer.valueDescriptor, flags.ValueEncoding)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func encodeProtobuf(data []byte, messageDescriptor *desc.MessageDescriptor, encoding string) (sarama.ByteEncoder, error) {
	data, err := decodeBytes(data, encoding)
	if err != nil {
		return nil, err
	}

	if messageDescriptor == nil {
		return data, nil
	}

	message := dynamic.NewMessage(messageDescriptor)
	if err = message.UnmarshalJSONPB(&jsonpb.Unmarshaler{AllowUnknownFields: true}, data); err != nil {
		return nil, errors.Wrap(err, "invalid json")
	}

	pb, err := message.Marshal()
	if err != nil {
		return nil, err
	}

	return pb, nil
}
