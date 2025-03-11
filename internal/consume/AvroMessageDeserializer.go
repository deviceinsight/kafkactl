package consume

import (
	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"github.com/riferrei/srclient"
)

type AvroMessageDeserializer struct {
	topic    string
	registry *internal.CachingSchemaRegistry
}

func (deserializer *AvroMessageDeserializer) canDeserialize(consumerMsg *sarama.ConsumerMessage, data []byte) bool {

	schemaID, err := deserializer.registry.ExtractSchemaID(data)
	if err == nil {
		schema, schemaErr := deserializer.registry.GetSchemaByID(schemaID)
		if schemaErr != nil {
			output.Debugf("schema not found. id=%d partition=%d, offset=%d error=%v", schemaID,
				consumerMsg.Partition, consumerMsg.Offset, err)
			return false
		}
		isAvroSchema := schema.SchemaType() == nil || *schema.SchemaType() == srclient.Avro
		return isAvroSchema
	}

	output.Debugf("failed to extract schema id. partition=%d, offset=%d error=%v", consumerMsg.Partition,
		consumerMsg.Offset, err)
	return false
}

func (deserializer *AvroMessageDeserializer) CanDeserializeKey(consumerMsg *sarama.ConsumerMessage,
	flags Flags) bool {
	return flags.KeyProtoType == "" && deserializer.canDeserialize(consumerMsg, consumerMsg.Key)
}

func (deserializer *AvroMessageDeserializer) CanDeserializeValue(consumerMsg *sarama.ConsumerMessage,
	flags Flags) bool {
	return flags.ValueProtoType == "" && deserializer.canDeserialize(consumerMsg, consumerMsg.Value)
}

func (deserializer *AvroMessageDeserializer) deserialize(data []byte) (*DeserializedData, error) {

	schemaID, err := deserializer.registry.ExtractSchemaID(data)
	if err != nil {
		return nil, err
	}

	schema, err := deserializer.registry.GetSchemaByID(schemaID)
	if err != nil {
		return nil, err
	}

	payload := deserializer.registry.ExtractPayload(data)

	output.Debugf("decode with schema id %d", schemaID)

	native, _, err := schema.Codec().NativeFromBinary(payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse avro data")
	}

	textual, err := schema.Codec().TextualFromNative(nil, native)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert data to avro data")
	}

	return &DeserializedData{schema: schema.Schema(), schemaID: &schemaID, data: textual}, nil
}

func (deserializer *AvroMessageDeserializer) DeserializeKey(consumerMsg *sarama.ConsumerMessage) (*DeserializedData, error) {
	output.Debugf("deserialize key with AvroMessageDeserializer")
	return deserializer.deserialize(consumerMsg.Key)
}

func (deserializer *AvroMessageDeserializer) DeserializeValue(consumerMsg *sarama.ConsumerMessage) (*DeserializedData, error) {
	output.Debugf("deserialize value with AvroMessageDeserializer")
	return deserializer.deserialize(consumerMsg.Value)
}
