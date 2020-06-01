package producer

import (
	"github.com/Shopify/sarama"
)

type MessageSerializer interface {
	Serialize(key, value []byte, flags ProducerFlags) (*sarama.ProducerMessage, error)
}
