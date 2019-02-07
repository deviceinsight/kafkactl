package consumer

import (
	"encoding/binary"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/landoop/schema-registry"
	"github.com/linkedin/goavro"
	"strconv"
	"strings"
	"time"
)

type AvroMessageDeserializer struct {
	topic              string
	avroSchemaRegistry string
	client             *schemaregistry.Client
}

func CreateAvroMessageDeserializer(topic string, avroSchemaRegistry string) AvroMessageDeserializer {

	var err error

	deserializer := AvroMessageDeserializer{topic: topic, avroSchemaRegistry: avroSchemaRegistry}

	deserializer.client, err = schemaregistry.NewClient(deserializer.avroSchemaRegistry)

	if err != nil {
		output.Failf("failed to create schema registry client: %v", err)
	}

	return deserializer
}

type avroMessage struct {
	Partition     int32
	Offset        int64
	KeySchema     *string `json:"keySchema,omitempty" yaml:"keySchema,omitempty"`
	KeySchemaId   *int    `json:"keySchemaId,omitempty" yaml:"keySchemaId,omitempty"`
	Key           *string `json:",omitempty" yaml:",omitempty"`
	ValueSchema   *string `json:"valueSchema,omitempty" yaml:"valueSchema,omitempty"`
	ValueSchemaId *int    `json:"valueSchemaId,omitempty" yaml:"valueSchemaId,omitempty"`
	Value         *string
	Timestamp     *time.Time `json:",omitempty" yaml:",omitempty"`
}

type decodedValue struct {
	schema   string
	schemaId int
	value    *string
}

func (deserializer AvroMessageDeserializer) newAvroMessage(msg *sarama.ConsumerMessage, flags ConsumerFlags) avroMessage {

	var decodedKey decodedValue
	var timestamp *time.Time

	var decodedValue = deserializer.decode(msg.Value, flags, "value")

	avroMessage := avroMessage{
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Value:     decodedValue.value,
		Timestamp: timestamp,
	}

	if flags.PrintKeys {
		decodedKey = deserializer.decode(msg.Key, flags, "key")
		avroMessage.Key = decodedKey.value
	}

	if flags.PrintTimestamps && !msg.Timestamp.IsZero() {
		avroMessage.Timestamp = &msg.Timestamp
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

	return avroMessage
}

func (deserializer AvroMessageDeserializer) decode(rawData []byte, flags ConsumerFlags, avroSchemaType string) decodedValue {

	if len(rawData) < 5 {
		// does not seem to be avro data
		return decodedValue{value: encodeBytes(rawData, flags.EncodeValue)}
	}

	// https://docs.confluent.io/current/schema-registry/docs/serializer-formatter.html#wire-format
	schemaId := int(binary.BigEndian.Uint32(rawData[1:5]))
	data := rawData[5:]
	subject := deserializer.topic + "-" + avroSchemaType

	schema, err := deserializer.client.GetSchemaBySubject(subject, schemaId)

	if err != nil {
		output.Failf("failed to find avro schema for subject: %s id: %d (%v)", subject, schemaId, err)
	}

	avroCodec, err := goavro.NewCodec(schema.Schema)

	if err != nil {
		output.Failf("failed to parse avro schema %v", err)
	}

	native, _, err := avroCodec.NativeFromBinary(data)
	if err != nil {
		output.Failf("failed to parse avro data: %s", err)
	}

	textual, err := avroCodec.TextualFromNative(nil, native)
	if err != nil {
		output.Failf("failed to convert value to avro data: %s", err)
	}

	decoded := string(textual)
	return decodedValue{schema: schema.Schema, schemaId: schema.ID, value: &decoded}
}

func (deserializer AvroMessageDeserializer) Deserialize(rawMsg *sarama.ConsumerMessage, flags ConsumerFlags) {

	msg := deserializer.newAvroMessage(rawMsg, flags)

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

		row = append(row, *msg.Value)

		output.PrintStrings(strings.Join(row[:], "#"))

	} else {
		output.PrintObject(msg, flags.OutputFormat)
	}

}
