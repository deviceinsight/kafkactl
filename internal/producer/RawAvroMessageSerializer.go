package producer

import (
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

// RawAvroMessageSerializer encodes message values as unframed Avro binary.
type RawAvroMessageSerializer struct {
	codec *avro.MessageCodec
}

func NewRawAvroMessageSerializer(codec *avro.MessageCodec) RawAvroMessageSerializer {
	return RawAvroMessageSerializer{codec: codec}
}

func (serializer RawAvroMessageSerializer) CanSerializeValue(_ string) (bool, error) {
	return true, nil
}

func (serializer RawAvroMessageSerializer) CanSerializeKey(_ string) (bool, error) {
	return false, nil
}

func (serializer RawAvroMessageSerializer) SerializeValue(value []byte, _ Flags) ([]byte, error) {
	// Only nil represents a Kafka tombstone. Pass non-nil input, including an
	// empty slice, to the Avro codec for validation.
	if value == nil {
		return nil, nil
	}

	output.Debugf("serialize value with RawAvroMessageSerializer")
	return serializer.codec.EncodeTextual(value)
}

func (serializer RawAvroMessageSerializer) SerializeKey(key []byte, _ Flags) ([]byte, error) {
	return key, nil
}
