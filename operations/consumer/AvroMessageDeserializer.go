package consumer

import (
	"encoding/binary"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/landoop/schema-registry"
	"github.com/linkedin/goavro"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type AvroMessageDeserializer struct {
	topic              string
	avroSchemaRegistry string
	client             *schemaregistry.Client
}

func CreateAvroMessageDeserializer(topic string, avroSchemaRegistry string) (AvroMessageDeserializer, error) {

	var err error

	deserializer := AvroMessageDeserializer{topic: topic, avroSchemaRegistry: avroSchemaRegistry}

	deserializer.client, err = schemaregistry.NewClient(deserializer.avroSchemaRegistry)

	if err != nil {
		return deserializer, errors.Wrap(err, "failed to create schema registry client: ")
	}

	return deserializer, nil
}

type avroMessage struct {
	Partition     int32
	Offset        int64
	Headers       map[string]string `json:",omitempty" yaml:",omitempty"`
	KeySchema     *string           `json:"keySchema,omitempty" yaml:"keySchema,omitempty"`
	KeySchemaId   *int              `json:"keySchemaId,omitempty" yaml:"keySchemaId,omitempty"`
	Key           *string           `json:",omitempty" yaml:",omitempty"`
	ValueSchema   *string           `json:"valueSchema,omitempty" yaml:"valueSchema,omitempty"`
	ValueSchemaId *int              `json:"valueSchemaId,omitempty" yaml:"valueSchemaId,omitempty"`
	Value         *string
	Timestamp     *time.Time `json:",omitempty" yaml:",omitempty"`
}

type decodedValue struct {
	schema   string
	schemaId int
	value    *string
}

func (deserializer AvroMessageDeserializer) newAvroMessage(msg *sarama.ConsumerMessage, flags ConsumerFlags) (avroMessage, error) {

	var decodedKey decodedValue
	var timestamp *time.Time

	decodedValue, err := deserializer.decode(msg.Value, flags, "value")
	if err != nil {
		return avroMessage{}, err
	}

	avroMessage := avroMessage{
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Value:     decodedValue.value,
		Timestamp: timestamp,
	}

	if flags.PrintKeys {
		decodedKey, err = deserializer.decode(msg.Key, flags, "key")
		if err != nil {
			return avroMessage, err
		}
		avroMessage.Key = decodedKey.value
	}

	if flags.PrintTimestamps && !msg.Timestamp.IsZero() {
		avroMessage.Timestamp = &msg.Timestamp
	}

	if flags.PrintHeaders {
		avroMessage.Headers = encodeRecordHeaders(msg.Headers)
	}

	if flags.PrintAvroSchema {
		if decodedKey.schemaId > 0 {
			avroMessage.KeySchema = &decodedKey.schema
			avroMessage.KeySchemaId = &decodedKey.schemaId
		}
		if decodedValue.schemaId > 0 {
			avroMessage.ValueSchema = &decodedValue.schema
			avroMessage.ValueSchemaId = &decodedValue.schemaId
		}
	}

	return avroMessage, nil
}

func (deserializer AvroMessageDeserializer) decode(rawData []byte, flags ConsumerFlags, avroSchemaType string) (decodedValue, error) {

	if len(rawData) < 5 {
		output.Debugf("does not seem to be avro data")
		return decodedValue{value: encodeBytes(rawData, flags.EncodeValue)}, nil
	}

	// https://docs.confluent.io/current/schema-registry/docs/serializer-formatter.html#wire-format
	schemaId := int(binary.BigEndian.Uint32(rawData[1:5]))
	data := rawData[5:]
	subject := deserializer.topic + "-" + avroSchemaType

	output.Debugf("decode %s and id %d", subject, schemaId)

	subjects, err := deserializer.client.Subjects()

	if err != nil {
		return decodedValue{}, errors.Wrap(err, "failed to list available avro schemas")
	}

	if !util.ContainsString(subjects, subject) {
		// does not seem to be avro data
		return decodedValue{value: encodeBytes(rawData, flags.EncodeValue)}, nil
	}

	schema, err := deserializer.client.GetSchemaByID(schemaId)

	if err != nil {
		return decodedValue{}, errors.Errorf("failed to find avro schema for subject: %s id: %d (%v)", subject, schemaId, err)
	}

	avroCodec, err := goavro.NewCodec(schema)

	if err != nil {
		return decodedValue{}, errors.Wrap(err, "failed to parse avro schema")
	}

	native, _, err := avroCodec.NativeFromBinary(data)
	if err != nil {
		return decodedValue{}, errors.Wrap(err, "failed to parse avro data")
	}

	textual, err := avroCodec.TextualFromNative(nil, native)
	if err != nil {
		return decodedValue{}, errors.Wrap(err, "failed to convert value to avro data")
	}

	decoded := string(textual)
	return decodedValue{schema: schema, schemaId: schemaId, value: &decoded}, nil
}

func (deserializer AvroMessageDeserializer) Deserialize(rawMsg *sarama.ConsumerMessage, flags ConsumerFlags) error {

	output.Debugf("start to deserialize avro message...")

	msg, err := deserializer.newAvroMessage(rawMsg, flags)

	if err != nil {
		return err
	}

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
		if flags.PrintAvroSchema {
			if msg.KeySchemaId != nil {
				row = append(row, *msg.KeySchema)
				row = append(row, strconv.Itoa(*msg.KeySchemaId))
			} else {
				row = append(row, "", "")
			}
			if msg.ValueSchemaId != nil {
				row = append(row, *msg.ValueSchema)
				row = append(row, strconv.Itoa(*msg.ValueSchemaId))
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

	} else {
		return output.PrintObject(msg, flags.OutputFormat)
	}
}
