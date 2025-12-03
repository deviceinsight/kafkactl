package consume

import (
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

type message struct {
	Partition     int32
	Offset        int64
	Headers       map[string]string `json:",omitempty" yaml:",omitempty"`
	KeySchema     *string           `json:"keySchema,omitempty" yaml:"keySchema,omitempty"`
	KeySchemaID   *int              `json:"keySchemaId,omitempty" yaml:"keySchemaId,omitempty"`
	Key           *string           `json:",omitempty" yaml:",omitempty"`
	ValueSchema   *string           `json:"valueSchema,omitempty" yaml:"valueSchema,omitempty"`
	ValueSchemaID *int              `json:"valueSchemaId,omitempty" yaml:"valueSchemaId,omitempty"`
	Value         *string
	Timestamp     *time.Time `json:",omitempty" yaml:",omitempty"`
}

type DeserializedData struct {
	schema   string
	schemaID *int
	data     []byte
}

func newMessage(consumerMsg *sarama.ConsumerMessage, flags Flags, key, value *DeserializedData) *message {

	msg := message{
		Partition: consumerMsg.Partition,
		Offset:    consumerMsg.Offset,
		Value:     encodeBytes(value.data, flags.EncodeValue),
	}

	if flags.PrintAll || flags.PrintKeys {
		if key != nil {
			msg.Key = encodeBytes(key.data, flags.EncodeKey)
		}
	}

	if flags.PrintAll || flags.PrintTimestamps && !consumerMsg.Timestamp.IsZero() {
		msg.Timestamp = &consumerMsg.Timestamp
	}

	if flags.PrintAll || flags.PrintHeaders {
		msg.Headers = encodeRecordHeaders(consumerMsg.Headers)
	}

	if flags.PrintAll || flags.PrintSchema {
		if key != nil && key.schemaID != nil {
			msg.KeySchema = &key.schema
			msg.KeySchemaID = key.schemaID
		}
		if value.schemaID != nil {
			msg.ValueSchema = &value.schema
			msg.ValueSchemaID = value.schemaID
		}
	}

	return &msg
}

func printMessage(msg *message, flags Flags) error {

	if flags.OutputFormat == "" {
		var row []string

		if flags.PrintAll || flags.PrintHeaders {
			if msg.Headers != nil {
				column := toSortedArray(msg.Headers)
				row = append(row, strings.Join(column[:], ","))
			} else {
				row = append(row, "")
			}
		}
		if flags.PrintAll || flags.PrintPartitions {
			row = append(row, strconv.Itoa(int(msg.Partition)))
		}
		if flags.PrintAll || flags.PrintKeys {
			if msg.Key != nil {
				row = append(row, *msg.Key)
			} else {
				row = append(row, "")
			}
		}
		if flags.PrintAll || flags.PrintTimestamps {
			if msg.Timestamp != nil {
				row = append(row, (*msg.Timestamp).Format(time.RFC3339))
			} else {
				row = append(row, "")
			}
		}
		if flags.PrintAll || flags.PrintSchema {
			if msg.KeySchemaID != nil {
				row = append(row, *msg.KeySchema)
				row = append(row, strconv.Itoa(*msg.KeySchemaID))
			} else {
				row = append(row, "", "")
			}
			if msg.ValueSchemaID != nil {
				row = append(row, *msg.ValueSchema)
				row = append(row, strconv.Itoa(*msg.ValueSchemaID))
			} else {
				row = append(row, "", "")
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
	}
	return output.PrintObject(msg, flags.OutputFormat)
}
