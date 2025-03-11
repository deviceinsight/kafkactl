package consume

import (
	"github.com/IBM/sarama"
)

type DefaultMessageDeserializer struct {
}

func (deserializer *DefaultMessageDeserializer) CanDeserializeKey(_ *sarama.ConsumerMessage, _ Flags) bool {
	return true
}

func (deserializer *DefaultMessageDeserializer) CanDeserializeValue(_ *sarama.ConsumerMessage, _ Flags) bool {
	return true
}

func (deserializer *DefaultMessageDeserializer) DeserializeKey(msg *sarama.ConsumerMessage) (*DeserializedData, error) {

	return &DeserializedData{
		schema:   "",
		schemaID: nil,
		data:     msg.Key,
	}, nil
}

func (deserializer *DefaultMessageDeserializer) DeserializeValue(msg *sarama.ConsumerMessage) (*DeserializedData, error) {

	return &DeserializedData{
		schema:   "",
		schemaID: nil,
		data:     msg.Value,
	}, nil
}
