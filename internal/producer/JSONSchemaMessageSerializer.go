package producer

import (
	"encoding/binary"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/riferrei/srclient"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

type JSONSchemaMessageSerializer struct {
	topic  string
	client *internal.CachingSchemaRegistry
}

func (serializer JSONSchemaMessageSerializer) CanSerializeValue(topic string) (bool, error) {
	return serializer.client.SubjectOfTypeExists(topic+"-value", srclient.Json)
}

func (serializer JSONSchemaMessageSerializer) CanSerializeKey(topic string) (bool, error) {
	return serializer.client.SubjectOfTypeExists(topic+"-key", srclient.Json)
}

func (serializer JSONSchemaMessageSerializer) SerializeValue(value []byte, flags Flags) ([]byte, error) {
	output.Debugf("serialize value with JSONSchemaMessageSerializer")
	return serializer.encode(value, flags.ValueSchemaVersion, serializer.topic+"-value")
}

func (serializer JSONSchemaMessageSerializer) SerializeKey(key []byte, flags Flags) ([]byte, error) {
	output.Debugf("serialize key with JSONSchemaMessageSerializer")
	return serializer.encode(key, flags.KeySchemaVersion, serializer.topic+"-key")
}

func (serializer JSONSchemaMessageSerializer) encode(rawData []byte, schemaVersion int, subject string) ([]byte, error) {

	var schema *srclient.Schema
	var err error

	if schemaVersion == -1 {
		schema, err = serializer.client.GetLatestSchema(subject)
		if err != nil {
			return nil, errors.Errorf("failed to find latest json schema for subject: %s (%v)", subject, err)
		}
		output.Debugf("encode with latest json schema id=%d subject=%s", schema.ID(), subject)
	} else {
		schema, err = serializer.client.GetSchemaByVersion(subject, schemaVersion)
		if err != nil {
			return nil, errors.Errorf("failed to find json schema for subject: %s version: %d (%v)", subject, schemaVersion, err)
		}
		output.Debugf("encode with json schema id=%d version=%d subject=%s", schema.ID(), schemaVersion, subject)
	}

	jsonSchema := schema.JsonSchema()
	if jsonSchema != nil {
		var v interface{}
		if err := json.Unmarshal(rawData, &v); err != nil {
			return nil, errors.Wrap(err, "failed to parse json data")
		}
		if err := jsonSchema.Validate(v); err != nil {
			return nil, errors.Wrap(err, "json data does not match schema")
		}
	}

	// https://docs.confluent.io/current/schema-registry/docs/serializer-formatter.html#wire-format
	versionBytes := make([]byte, 5)
	binary.BigEndian.PutUint32(versionBytes[1:], uint32(schema.ID()))

	return append(versionBytes, rawData...), nil
}
