package consumer

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"strings"
	"time"
)

type DefaultMessageDeserializer struct {
}

type defaultMessage struct {
	Partition int32
	Offset    int64
	Headers   map[string]string `json:",omitempty" yaml:",omitempty"`
	Key       *string           `json:",omitempty" yaml:",omitempty"`
	Value     *string
	Timestamp *time.Time `json:",omitempty" yaml:",omitempty"`
}

func (deserializer DefaultMessageDeserializer) newDefaultMessage(msg *sarama.ConsumerMessage, flags ConsumerFlags) defaultMessage {

	var key *string
	var timestamp *time.Time
	var headers map[string]string

	var value = encodeBytes(msg.Value, flags.EncodeValue)

	if flags.PrintKeys {
		key = encodeBytes(msg.Key, flags.EncodeKey)
	}

	if flags.PrintTimestamps && !msg.Timestamp.IsZero() {
		timestamp = &msg.Timestamp
	}

	if flags.PrintHeaders {
		headers = encodeRecordHeaders(msg.Headers)
	}

	return defaultMessage{
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Headers:   headers,
		Key:       key,
		Value:     value,
		Timestamp: timestamp,
	}
}

func (deserializer DefaultMessageDeserializer) Deserialize(rawMsg *sarama.ConsumerMessage, flags ConsumerFlags) error {

	msg := deserializer.newDefaultMessage(rawMsg, flags)

	if flags.OutputFormat == "" {
		var row []string

		if flags.PrintHeaders {
			if msg.Headers != nil {
				var column []string

				for key, value := range msg.Headers {
					column = append(column, key+":"+value)
				}

				row = append(row, strings.Join(column[:], ","))
			} else {
				row = append(row, "")
			}
		}

		if flags.PrintKeys {
			if msg.Key != nil {
				row = append(row, *msg.Key)
			} else {
				row = append(row, "")
			}
		}
		if flags.PrintTimestamps {
			if msg.Timestamp != nil {
				row = append(row, (*msg.Timestamp).Format(time.RFC3339))
			} else {
				row = append(row, "")
			}
		}

		var value string

		if msg.Value != nil {
			value = *msg.Value
		} else {
			value = "null"
		}

		row = append(row, value)

		output.PrintStrings(strings.Join(row[:], flags.Separator))
		return nil
	} else {
		return output.PrintObject(msg, flags.OutputFormat)
	}
}
