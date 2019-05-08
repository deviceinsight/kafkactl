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
	Key       *string `json:",omitempty" yaml:",omitempty"`
	Value     *string
	Timestamp *time.Time `json:",omitempty" yaml:",omitempty"`
}

func (deserializer DefaultMessageDeserializer) newDefaultMessage(msg *sarama.ConsumerMessage, flags ConsumerFlags) defaultMessage {

	var key *string
	var timestamp *time.Time

	var value = encodeBytes(msg.Value, flags.EncodeValue)

	if flags.PrintKeys {
		key = encodeBytes(msg.Key, flags.EncodeKey)
	}

	if flags.PrintTimestamps && !msg.Timestamp.IsZero() {
		timestamp = &msg.Timestamp
	}

	return defaultMessage{
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Key:       key,
		Value:     value,
		Timestamp: timestamp,
	}
}

func (deserializer DefaultMessageDeserializer) Deserialize(rawMsg *sarama.ConsumerMessage, flags ConsumerFlags) {

	msg := deserializer.newDefaultMessage(rawMsg, flags)

	if flags.OutputFormat == "" {
		var row []string

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

		row = append(row, string(value))

		output.PrintStrings(strings.Join(row[:], "#"))

	} else {
		output.PrintObject(msg, flags.OutputFormat)
	}
}
