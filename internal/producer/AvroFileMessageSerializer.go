package producer

import (
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

// AvroFileMessageSerializer encodes avro values with --avro-schema-file, without the Confluent wire-format prefix.
type AvroFileMessageSerializer struct {
	codec *goavro.Codec
}

func CreateAvroFileMessageSerializer(location string, source *avro.SchemaSource) (*AvroFileMessageSerializer, error) {
	codec, err := source.Resolve(location)
	if err != nil {
		return nil, err
	}

	return &AvroFileMessageSerializer{codec: codec}, nil
}

func (serializer AvroFileMessageSerializer) CanSerializeValue(_ string) (bool, error) {
	return true, nil
}

func (serializer AvroFileMessageSerializer) CanSerializeKey(_ string) (bool, error) {
	return false, nil
}

func (serializer AvroFileMessageSerializer) SerializeValue(value []byte, _ Flags) ([]byte, error) {
	output.Debugf("serialize value with AvroFileMessageSerializer")

	native, _, err := serializer.codec.NativeFromTextual(value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert value to avro data")
	}

	data, err := serializer.codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert value to avro data")
	}

	return data, nil
}

func (serializer AvroFileMessageSerializer) SerializeKey(_ []byte, _ Flags) ([]byte, error) {
	return nil, errors.New("AvroFileMessageSerializer does not support keys")
}
