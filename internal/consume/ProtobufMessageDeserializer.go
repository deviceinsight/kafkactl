package consume

import (
	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

type ProtobufMessageDeserializer struct {
	keyType         string
	valueType       string
	keyDescriptor   *desc.MessageDescriptor
	valueDescriptor *desc.MessageDescriptor
	jsonMarshaler   *jsonpb.Marshaler
}

func CreateProtobufMessageDeserializer(context internal.ProtobufConfig, keyType, valueType string, jsonMarshaler *jsonpb.Marshaler) (*ProtobufMessageDeserializer, error) {
	return &ProtobufMessageDeserializer{
		keyType:         keyType,
		valueType:       valueType,
		keyDescriptor:   protobuf.ResolveMessageType(context, keyType),
		valueDescriptor: protobuf.ResolveMessageType(context, valueType),
		jsonMarshaler:   jsonMarshaler,
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

	deserialized, err := decodeProtobuf(consumerMsg.Key, deserializer.keyDescriptor, deserializer.jsonMarshaler)
	return &DeserializedData{data: deserialized}, err
}

func (deserializer *ProtobufMessageDeserializer) DeserializeValue(consumerMsg *sarama.ConsumerMessage) (*DeserializedData, error) {
	output.Debugf("deserialize value with ProtobufMessageDeserializer")
	if deserializer.valueDescriptor == nil {
		return nil, errors.Errorf("value message type %q not found in provided files", deserializer.valueType)
	}
	deserialized, err := decodeProtobuf(consumerMsg.Value, deserializer.valueDescriptor, deserializer.jsonMarshaler)
	return &DeserializedData{data: deserialized}, err
}

func decodeProtobuf(b []byte, msgDesc *desc.MessageDescriptor, jsonMarshaler *jsonpb.Marshaler) ([]byte, error) {
	if len(b) == 0 { // tombstone record, can't be unmarshalled as protobuf
		return nil, nil
	}

	msg := dynamic.NewMessage(msgDesc)
	if err := msg.Unmarshal(b); err != nil {
		return nil, err
	}

	j, err := msg.MarshalJSONPB(jsonMarshaler)
	if err != nil {
		return nil, err
	}

	return j, nil
}
