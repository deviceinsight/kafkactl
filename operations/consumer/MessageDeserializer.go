package consumer

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/Shopify/sarama"
)

type MessageDeserializer interface {
	Deserialize(msg *sarama.ConsumerMessage, flags ConsumerFlags) error
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

func encodeRecordHeaders(headers []*sarama.RecordHeader) map[string]string {
	if headers == nil {
		return nil
	}

	data := make(map[string]string)

	for _, header := range headers {
		if header.Key == nil {
			continue
		}

		key := *encodeBytes(header.Key, "")
		value := *encodeBytes(header.Value, "")

		data[key] = value
	}

	return data
}
