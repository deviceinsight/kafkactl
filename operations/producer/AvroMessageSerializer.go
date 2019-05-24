package producer

import (
	"encoding/binary"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/landoop/schema-registry"
	"github.com/linkedin/goavro"
)

type AvroMessageSerializer struct {
	topic              string
	avroSchemaRegistry string
	client             *schemaregistry.Client
}

func CreateAvroMessageSerializer(topic string, avroSchemaRegistry string) AvroMessageSerializer {

	var err error

	serializer := AvroMessageSerializer{topic: topic, avroSchemaRegistry: avroSchemaRegistry}

	serializer.client, err = schemaregistry.NewClient(serializer.avroSchemaRegistry)

	if err != nil {
		output.Failf("failed to create schema registry client: %v", err)
	}

	return serializer
}

func (serializer AvroMessageSerializer) encode(rawData []byte, schemaVersion int, avroSchemaType string) []byte {

	subject := serializer.topic + "-" + avroSchemaType

	subjects, err := serializer.client.Subjects()

	if err != nil {
		output.Failf("failed to list available avro schemas (%v)", err)
	}

	if !util.ContainsString(subjects, subject) {
		// does not seem to be avro data
		return rawData
	}

	var schema schemaregistry.Schema

	if schemaVersion == -1 {
		schema, err = serializer.client.GetLatestSchema(subject)

		if err != nil {
			output.Failf("failed to find latest avro schema for subject: %s (%v)", subject, err)
		}
	} else {
		schema, err = serializer.client.GetSchemaBySubject(subject, schemaVersion)

		if err != nil {
			output.Failf("failed to find avro schema for subject: %s id: %d (%v)", subject, schemaVersion, err)
		}
	}

	codec, err := goavro.NewCodec(schema.Schema)

	if err != nil {
		output.Failf("failed to parse avro schema: %s", err)
	}

	native, _, err := codec.NativeFromTextual(rawData)
	if err != nil {
		output.Failf("failed to convert value to avro data: %s", err)
	}

	data, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		output.Failf("failed to convert value to avro data: %s", err)
	}

	// https://docs.confluent.io/current/schema-registry/docs/serializer-formatter.html#wire-format
	versionBytes := make([]byte, 5)
	binary.BigEndian.PutUint32(versionBytes[1:], uint32(schema.ID))

	return append(versionBytes, data...)
}

func (serializer AvroMessageSerializer) Serialize(key, value []byte, flags ProducerFlags) *sarama.ProducerMessage {

	message := &sarama.ProducerMessage{Topic: serializer.topic, Partition: flags.Partition}

	if key != nil {
		message.Key = sarama.ByteEncoder(serializer.encode(value, flags.KeySchemaVersion, "key"))
	}

	message.Value = sarama.ByteEncoder(serializer.encode(value, flags.ValueSchemaVersion, "value"))

	return message
}
