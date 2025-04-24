package producer

import (
	"encoding/binary"
	"fmt"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"
	"github.com/pkg/errors"
	"github.com/riferrei/srclient"
)

type RegistryProtobufMessageSerializer struct {
	topic  string
	client *internal.CachingSchemaRegistry
}

func (serializer RegistryProtobufMessageSerializer) CanSerializeValue(topic string) (bool, error) {
	return serializer.client.SubjectOfTypeExists(topic+"-value", srclient.Protobuf)
}

func (serializer RegistryProtobufMessageSerializer) CanSerializeKey(topic string) (bool, error) {
	return serializer.client.SubjectOfTypeExists(topic+"-key", srclient.Protobuf)
}

func (serializer RegistryProtobufMessageSerializer) SerializeValue(value []byte, flags Flags) ([]byte, error) {
	return serializer.encode(value, flags.ValueSchemaVersion, flags.ValueProtoType, serializer.topic+"-value")
}

func (serializer RegistryProtobufMessageSerializer) SerializeKey(key []byte, flags Flags) ([]byte, error) {
	return serializer.encode(key, flags.KeySchemaVersion, flags.KeyProtoType, serializer.topic+"-key")
}

func (serializer RegistryProtobufMessageSerializer) encode(rawData []byte, schemaVersion int, msgName string, subject string) ([]byte, error) {
	var schema *srclient.Schema
	var err error

	if schemaVersion == -1 {
		schema, err = serializer.client.GetLatestSchema(subject)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("couldn't get schema for subject %s", subject))
		}
	} else {
		schema, err = serializer.client.GetSchema(schemaVersion)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("couldn't get schema for schemaId %d", schemaVersion))
		}
	}

	fileDesc, err := protobuf.SchemaToFileDescriptor(serializer.client, schema)
	if err != nil {
		return nil, err
	}

	messageDesc := fileDesc.FindMessage(msgName)
	if messageDesc == nil {
		messageDesc = fileDesc.GetMessageTypes()[0]
		msgName = messageDesc.GetFullyQualifiedName()
	}

	pb, err := encodeProtobuf(rawData, messageDesc)
	if err != nil {
		return nil, err
	}

	result := []byte{internal.MagicByte}
	result = binary.BigEndian.AppendUint32(result, uint32(schema.ID()))
	indexes, err := protobuf.ComputeIndexes(fileDesc, msgName)
	if err != nil {
		return nil, err
	}

	result = binary.AppendVarint(result, int64(len(indexes)))
	if len(indexes) != 0 {
		for _, idx := range indexes {
			result = binary.AppendVarint(result, idx)
		}
	}

	result = append(result, pb...)

	return result, nil
}
