package producer

import (
	"encoding/binary"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/riferrei/srclient"

	"github.com/deviceinsight/kafkactl/v5/internal/util"
	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

type AvroMessageSerializer struct {
	topic     string
	jsonCodec avro.JSONCodec
	client    *internal.CachingSchemaRegistry
}

func (serializer AvroMessageSerializer) CanSerializeValue(topic string) (bool, error) {
	return serializer.client.SubjectOfTypeExists(topic+"-value", srclient.Avro)
}

func (serializer AvroMessageSerializer) CanSerializeKey(topic string) (bool, error) {
	return serializer.client.SubjectOfTypeExists(topic+"-key", srclient.Avro)
}

func (serializer AvroMessageSerializer) SerializeValue(value []byte, flags Flags) ([]byte, error) {
	return serializer.encode(value, flags.ValueSchemaVersion, serializer.topic+"-value")
}

func (serializer AvroMessageSerializer) SerializeKey(key []byte, flags Flags) ([]byte, error) {
	return serializer.encode(key, flags.KeySchemaVersion, serializer.topic+"-key")
}

func (serializer AvroMessageSerializer) encode(rawData []byte, schemaVersion int, subject string) ([]byte, error) {

	var schema *srclient.Schema
	var err error

	if schemaVersion == -1 {
		schema, err = serializer.client.GetLatestSchema(subject)
		if err != nil {
			return nil, errors.Errorf("failed to find latest avro schema for subject: %s (%v)", subject, err)
		}
	} else {
		schema, err = serializer.client.GetSchemaByVersion(subject, schemaVersion)
		if err != nil {
			return nil, errors.Errorf("failed to find avro schema for subject: %s id: %d (%v)", subject, schemaVersion, err)
		}
	}

	var avroCodec *goavro.Codec

	if serializer.jsonCodec == avro.Avro {
		avroCodec, err = goavro.NewCodec(schema.Schema())
	} else {
		avroCodec, err = goavro.NewCodecForStandardJSONFull(schema.Schema())
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
	binary.BigEndian.PutUint32(versionBytes[1:], uint32(schema.ID()))

	return append(versionBytes, data...), nil
}

func (serializer AvroMessageSerializer) CanSerialize(topic string) (bool, error) {
	subjects, err := serializer.client.GetSubjects()
	if err != nil {
		return false, errors.Wrap(err, "failed to list available avro schemas")
	}

	// we currently only implement the TopicNameStrategy to map from a message to the corresponding schema
	// in this strategy the name of the schema subject is derived from topic name i.e. `topic+"-key"`, `topic+"-value"`
	if util.ContainsString(subjects, topic+"-key") {
		return true, nil
	} else if util.ContainsString(subjects, topic+"-value") {
		return true, nil
	}
	return false, nil
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
