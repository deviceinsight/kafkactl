package consumer

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/Shopify/sarama"
)

type MessageDeserializer interface {
	Deserialize(msg *sarama.ConsumerMessage, flags ConsumerFlags)
}

func encodeBytes(data []byte, encoding string) *string {
	if data == nil {
		return nil
	}

	var str string
	switch encoding {
	case "hex":
		str = hex.EncodeToString(data)
	case "base64":
		str = base64.StdEncoding.EncodeToString(data)
	default:
		str = string(data)
	}

	return &str
}
