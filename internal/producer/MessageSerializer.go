package producer

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type messageSerializer interface {
	CanSerializeValue(topic string) (bool, error)
	CanSerializeKey(topic string) (bool, error)
	SerializeValue(value []byte, flags Flags) ([]byte, error)
	SerializeKey(key []byte, flags Flags) ([]byte, error)
}

const (
	HEX    = "hex"
	BASE64 = "base64"
)

func parseHeader(raw string) (key, value string, err error) {
	// use a regex to split in order to handle escaped colons
	re := regexp.MustCompile(`[^\\]:`)

	// regexp.Split() cannot be used because regexp doesn't support negative lookbehind
	index := re.FindAllStringIndex(raw, 1)

	if len(index) != 1 {
		return "", "", errors.Errorf("unable to parse header: %s", raw)
	}

	runes := []rune(raw)
	separatorIdx := index[0]
	key = string(runes[0 : separatorIdx[0]+1])
	value = string(runes[separatorIdx[1]:])

	return key, value, nil
}

func createRecordHeaders(flags Flags) ([]sarama.RecordHeader, error) {
	recordHeaders := make([]sarama.RecordHeader, len(flags.Headers))

	for i, header := range flags.Headers {
		key, value, err := parseHeader(header)
		if err != nil {
			return nil, err
		}

		recordHeaders[i].Key = sarama.ByteEncoder(key)
		recordHeaders[i].Value = sarama.ByteEncoder(value)
	}
	return recordHeaders, nil
}

func decodeBytes(data []byte, encoding string) ([]byte, error) {
	if data == nil {
		return nil, nil
	}

	var out []byte
	switch encoding {
	case HEX:
		out = make([]byte, hex.DecodedLen(len(data)))
		bytelen, err := hex.Decode(out, data)
		if err != nil {
			return nil, err
		}
		return out[:bytelen], nil
	case BASE64:
		out = make([]byte, base64.StdEncoding.DecodedLen(len(data)))
		bytelen, err := base64.StdEncoding.Decode(out, data)
		if err != nil {
			return nil, err
		}
		return out[:bytelen], nil
	default:
		return data, nil
	}
}
