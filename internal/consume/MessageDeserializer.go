package consume

import (
	"encoding/base64"
	"encoding/hex"
	"sort"
	"unicode/utf8"

	"github.com/IBM/sarama"
)

type MessageDeserializer interface {
	CanDeserialize(topic string) (bool, error)
	Deserialize(msg *sarama.ConsumerMessage, flags Flags) error
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
		value := *encodeBytes(header.Value, "")

		data[key] = value
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
