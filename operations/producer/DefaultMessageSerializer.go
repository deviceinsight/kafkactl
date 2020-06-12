package producer

import (
	"github.com/Shopify/sarama"
)

type DefaultMessageSerializer struct {
	topic string
}

func (serializer DefaultMessageSerializer) Serialize(key, value []byte, flags ProducerFlags) (*sarama.ProducerMessage, error) {

	message := &sarama.ProducerMessage{Topic: serializer.topic, Partition: flags.Partition}

	if key != nil {
		message.Key = sarama.ByteEncoder(key)
	}

	message.Value = sarama.ByteEncoder(value)

	return message, nil
}
