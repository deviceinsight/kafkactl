package consume

import (
	"github.com/IBM/sarama"
	"github.com/riferrei/srclient"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

type JSONSchemaMessageDeserializer struct {
	topic    string
	registry *internal.CachingSchemaRegistry
}

func (deserializer *JSONSchemaMessageDeserializer) canDeserialize(consumerMsg *sarama.ConsumerMessage, data []byte) bool {
	schemaID, err := deserializer.registry.ExtractSchemaID(data)
	if err == nil {
		schema, schemaErr := deserializer.registry.GetSchema(schemaID)
		if schemaErr != nil {
			output.Debugf("schema not found. id=%d partition=%d, offset=%d error=%v", schemaID,
				consumerMsg.Partition, consumerMsg.Offset, schemaErr)
			return false
		}
		return schema.SchemaType() != nil && *schema.SchemaType() == srclient.Json
	}

	output.Debugf("failed to extract schema id. partition=%d, offset=%d error=%v", consumerMsg.Partition,
		consumerMsg.Offset, err)
	return false
}

func (deserializer *JSONSchemaMessageDeserializer) CanDeserializeKey(consumerMsg *sarama.ConsumerMessage,
	flags Flags,
) bool {
	return flags.KeyProtoType == "" && deserializer.canDeserialize(consumerMsg, consumerMsg.Key)
}

func (deserializer *JSONSchemaMessageDeserializer) CanDeserializeValue(consumerMsg *sarama.ConsumerMessage,
	flags Flags,
) bool {
	return flags.ValueProtoType == "" && deserializer.canDeserialize(consumerMsg, consumerMsg.Value)
}

func (deserializer *JSONSchemaMessageDeserializer) deserialize(data []byte) (*DeserializedData, error) {
	schemaID, err := deserializer.registry.ExtractSchemaID(data)
	if err != nil {
		return nil, err
	}

	schema, err := deserializer.registry.GetSchema(schemaID)
	if err != nil {
		return nil, err
	}

	payload := deserializer.registry.ExtractPayload(data)

	output.Debugf("decode with json schema id=%d", schemaID)

	return &DeserializedData{schema: schema.Schema(), schemaID: &schemaID, data: payload}, nil
}

func (deserializer *JSONSchemaMessageDeserializer) DeserializeKey(consumerMsg *sarama.ConsumerMessage) (*DeserializedData, error) {
	output.Debugf("deserialize key with JSONSchemaMessageDeserializer")
	return deserializer.deserialize(consumerMsg.Key)
}

func (deserializer *JSONSchemaMessageDeserializer) DeserializeValue(consumerMsg *sarama.ConsumerMessage) (*DeserializedData, error) {
	output.Debugf("deserialize value with JSONSchemaMessageDeserializer")
	return deserializer.deserialize(consumerMsg.Value)
}
