package consume

import (
	"encoding/base64"
	"encoding/hex"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/IBM/sarama"
)

type MessageDeserializer interface {
	CanDeserializeKey(msg *sarama.ConsumerMessage, flags Flags) bool
	CanDeserializeValue(msg *sarama.ConsumerMessage, flags Flags) bool
	DeserializeKey(msg *sarama.ConsumerMessage) (*DeserializedData, error)
	DeserializeValue(msg *sarama.ConsumerMessage) (*DeserializedData, error)
}

const (
	HEX    = "hex"
	BASE64 = "base64"
	NONE   = "none"
)

func encodeBytes(data []byte, encoding string) *string {
	if data == nil {
		return nil
	}

	if encoding != HEX && encoding != BASE64 && encoding != NONE {
		// Auto-detect encoding if no parameter is set
		if !utf8.Valid(data) {
			encoding = BASE64
		} else {
			encoding = NONE
		}
	}

	var str string
	switch encoding {
	case HEX:
		str = hex.EncodeToString(data)
	case BASE64:
		str = base64.StdEncoding.EncodeToString(data)
	case NONE:
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
		var value string
		if encodedValue := encodeBytes(decodeAMQPValue(header.Value), ""); encodedValue != nil {
			value = *encodedValue
		}

		data[key] = value
	}

	return data
}

// decodeAMQPValue attempts to decode AMQP-encoded header values.
// Azure Event Hubs wraps Kafka header values in AMQP type descriptors when using the Kafka interface.
// This not moved to the azure plugin, because of the required RPC overhead per header value.
// For regular Kafka servers, this function returns the input unchanged.
func decodeAMQPValue(data []byte) []byte {
	if len(data) < 2 {
		return data
	}

	switch data[0] {
	case 0xa1: // str8: string with 1-byte length
		length := int(data[1])
		if len(data) >= 2+length {
			return data[2 : 2+length]
		}
	case 0xb1: // str32: string with 4-byte length
		if len(data) >= 6 {
			length := int(data[1])<<24 | int(data[2])<<16 | int(data[3])<<8 | int(data[4])
			if len(data) >= 5+length {
				return data[5 : 5+length]
			}
		}
	case 0xa0: // vbin8: binary with 1-byte length
		length := int(data[1])
		if len(data) >= 2+length {
			return data[2 : 2+length]
		}
	case 0xb0: // vbin32: binary with 4-byte length
		if len(data) >= 6 {
			length := int(data[1])<<24 | int(data[2])<<16 | int(data[3])<<8 | int(data[4])
			if len(data) >= 5+length {
				return data[5 : 5+length]
			}
		}
	case 0x83: // timestamp: 64-bit milliseconds since Unix epoch
		if len(data) >= 9 {
			ms := int64(data[1])<<56 | int64(data[2])<<48 | int64(data[3])<<40 | int64(data[4])<<32 |
				int64(data[5])<<24 | int64(data[6])<<16 | int64(data[7])<<8 | int64(data[8])
			t := time.UnixMilli(ms)
			return []byte(t.Format(time.RFC3339))
		}
	}

	return data
}

func toSortedArray(headers map[string]string) []string {

	var column []string

	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		column = append(column, key+":"+headers[key])
	}
	return column
}
