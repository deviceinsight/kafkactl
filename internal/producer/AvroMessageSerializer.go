package producer

import (
	"encoding/binary"

	"github.com/deviceinsight/kafkactl/internal/helpers/avro"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/util"
	schemaregistry "github.com/landoop/schema-registry"
	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

type AvroMessageSerializer struct {
	topic              string
	avroSchemaRegistry string
	jsonCodec          avro.JSONCodec
	client             *schemaregistry.Client
}

func CreateAvroMessageSerializer(topic string, avroSchemaRegistry string, jsonCodec avro.JSONCodec) (AvroMessageSerializer, error) {

	var err error

	serializer := AvroMessageSerializer{topic: topic, avroSchemaRegistry: avroSchemaRegistry, jsonCodec: jsonCodec}

	serializer.client, err = schemaregistry.NewClient(serializer.avroSchemaRegistry)

	if err != nil {
		return serializer, errors.Wrap(err, "failed to create schema registry client: ")
	}

	return serializer, nil
}

func (serializer AvroMessageSerializer) encode(rawData []byte, schemaVersion int, avroSchemaType string) ([]byte, error) {

	subject := serializer.topic + "-" + avroSchemaType

	subjects, err := serializer.client.Subjects()

	if err != nil {
		return nil, errors.Wrap(err, "failed to list available avro schemas")
	}

	if !util.ContainsString(subjects, subject) {
		// does not seem to be avro data
		return rawData, nil
	}

	var schema schemaregistry.Schema

	if schemaVersion == -1 {
		schema, err = serializer.client.GetLatestSchema(subject)

		if err != nil {
			return nil, errors.Errorf("failed to find latest avro schema for subject: %s (%v)", subject, err)
		}
	} else {
		schema, err = serializer.client.GetSchemaBySubject(subject, schemaVersion)

		if err != nil {
			return nil, errors.Errorf("failed to find avro schema for subject: %s id: %d (%v)", subject, schemaVersion, err)
		}
	}

	var avroCodec *goavro.Codec

	if serializer.jsonCodec == avro.Avro {
		avroCodec, err = goavro.NewCodec(schema.Schema)
	} else {
		avroCodec, err = goavro.NewCodecForStandardJSONFull(schema.Schema)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse avro schema")
	}

	native, _, err := avroCodec.NativeFromTextual(rawData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert value to avro data")
	}

	data, err := avroCodec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert value to avro data")
	}

	// https://docs.confluent.io/current/schema-registry/docs/serializer-formatter.html#wire-format
	versionBytes := make([]byte, 5)
	binary.BigEndian.PutUint32(versionBytes[1:], uint32(schema.ID))

	return append(versionBytes, data...), nil
}

func (serializer AvroMessageSerializer) CanSerialize(topic string) (bool, error) {

	subjects, err := serializer.client.Subjects()

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

func (serializer AvroMessageSerializer) Serialize(key, value []byte, flags Flags) (*sarama.ProducerMessage, error) {

	recordHeaders, err := createRecordHeaders(flags)
	if err != nil {
		return nil, err
	}
	message := &sarama.ProducerMessage{Topic: serializer.topic, Partition: flags.Partition, Headers: recordHeaders}

	if key != nil {
		bytes, err := serializer.encode(key, flags.KeySchemaVersion, "key")
		if err != nil {
			return nil, err
		}
		message.Key = sarama.ByteEncoder(bytes)
	}

	bytes, err := serializer.encode(value, flags.ValueSchemaVersion, "value")
	if err != nil {
		return nil, err
	}
	message.Value = sarama.ByteEncoder(bytes)

	return message, nil
}
