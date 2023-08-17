package consume

import (
	"encoding/binary"
	"strconv"
	"strings"
	"time"

	"github.com/deviceinsight/kafkactl/internal/helpers/avro"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

type AvroMessageDeserializer struct {
	topic              string
	avroSchemaRegistry string
	jsonCodec          avro.JSONCodec
	registry           *CachingSchemaRegistry
}

func CreateAvroMessageDeserializer(topic string, avroSchemaRegistry string, jsonCodec avro.JSONCodec) (AvroMessageDeserializer, error) {

	var err error

	deserializer := AvroMessageDeserializer{topic: topic, avroSchemaRegistry: avroSchemaRegistry, jsonCodec: jsonCodec}

	deserializer.registry, err = CreateCachingSchemaRegistry(deserializer.avroSchemaRegistry)

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
	KeySchemaID   *int              `json:"keySchemaId,omitempty" yaml:"keySchemaId,omitempty"`
	Key           *string           `json:",omitempty" yaml:",omitempty"`
	ValueSchema   *string           `json:"valueSchema,omitempty" yaml:"valueSchema,omitempty"`
	ValueSchemaID *int              `json:"valueSchemaId,omitempty" yaml:"valueSchemaId,omitempty"`
	Value         *string
	Timestamp     *time.Time `json:",omitempty" yaml:",omitempty"`
}

type decodedValue struct {
	schema   string
	schemaID int
	value    *string
}

func (deserializer AvroMessageDeserializer) newAvroMessage(msg *sarama.ConsumerMessage, flags Flags) (avroMessage, error) {

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
		if decodedKey.schemaID > 0 {
			avroMessage.KeySchema = &decodedKey.schema
			avroMessage.KeySchemaID = &decodedKey.schemaID
		}
		if decodedValue.schemaID > 0 {
			avroMessage.ValueSchema = &decodedValue.schema
			avroMessage.ValueSchemaID = &decodedValue.schemaID
		}
	}

	return avroMessage, nil
}

func (deserializer AvroMessageDeserializer) decode(rawData []byte, flags Flags, avroSchemaType string) (decodedValue, error) {

	if len(rawData) < 5 {
		output.Debugf("does not seem to be avro data")
		return decodedValue{value: encodeBytes(rawData, flags.EncodeValue)}, nil
	}

	// https://docs.confluent.io/current/schema-registry/docs/serializer-formatter.html#wire-format
	schemaID := int(binary.BigEndian.Uint32(rawData[1:5]))
	data := rawData[5:]
	subject := deserializer.topic + "-" + avroSchemaType

	output.Debugf("decode %s and id %d", subject, schemaID)

	subjects, err := deserializer.registry.Subjects()

	if err != nil {
		return decodedValue{}, errors.Wrap(err, "failed to list available avro schemas")
	}

	if !util.ContainsString(subjects, subject) {
		// does not seem to be avro data
		return decodedValue{value: encodeBytes(rawData, flags.EncodeValue)}, nil
	}

	schema, err := deserializer.registry.GetSchemaByID(schemaID)

	if err != nil {
		return decodedValue{}, errors.Errorf("failed to find avro schema for subject: %s id: %d (%v)", subject, schemaID, err)
	}

	var avroCodec *goavro.Codec

	if deserializer.jsonCodec == avro.Avro {
		avroCodec, err = goavro.NewCodec(schema)
	} else {
		avroCodec, err = goavro.NewCodecForStandardJSONFull(schema)
	}

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
	return decodedValue{schema: schema, schemaID: schemaID, value: &decoded}, nil
}

func (deserializer AvroMessageDeserializer) CanDeserialize(topic string) (bool, error) {
	subjects, err := deserializer.registry.Subjects()

	if err != nil {
		return false, errors.Wrap(err, "failed to list available avro schemas")
	}

	if util.ContainsString(subjects, topic+"-key") {
		return true, nil
	} else if util.ContainsString(subjects, topic+"-value") {
		return true, nil
	} else {
		return false, nil
	}
}

func (deserializer AvroMessageDeserializer) Deserialize(rawMsg *sarama.ConsumerMessage, flags Flags) error {

	output.Debugf("start to deserialize avro message...")

	msg, err := deserializer.newAvroMessage(rawMsg, flags)

	if err != nil {
		return err
	}

	if flags.OutputFormat == "" {
		var row []string

		if flags.PrintHeaders {
			if msg.Headers != nil {
				column := toSortedArray(msg.Headers)
				row = append(row, strings.Join(column[:], ","))
			} else {
				row = append(row, "")
			}
		}
		if flags.PrintPartitions {
			row = append(row, strconv.Itoa(int(msg.Partition)))
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
