package consume

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/cloudevents"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

// AvroFileMessageDeserializer decodes avro values without a schema registry, using a static schema
// (--avro-schema-file) or a per message schema uri (--avro-schema-header, e.g. ce_dataschema).
type AvroFileMessageDeserializer struct {
	schemaHeader string
	staticCodec  *goavro.Codec
	source       *avro.SchemaSource
}

func CreateAvroFileMessageDeserializer(schemaFile, schemaHeader string, source *avro.SchemaSource) (*AvroFileMessageDeserializer, error) {
	deserializer := &AvroFileMessageDeserializer{schemaHeader: schemaHeader, source: source}

	if schemaFile != "" {
		codec, err := source.Resolve(schemaFile)
		if err != nil {
			return nil, err
		}
		deserializer.staticCodec = codec
	}

	return deserializer, nil
}

func (deserializer *AvroFileMessageDeserializer) CanDeserializeKey(_ *sarama.ConsumerMessage, _ Flags) bool {
	return false
}

func (deserializer *AvroFileMessageDeserializer) CanDeserializeValue(consumerMsg *sarama.ConsumerMessage, flags Flags) bool {
	if flags.ValueProtoType != "" {
		return false
	}

	switch contentModeOf(consumerMsg) {
	case cloudevents.StructuredJSONMode:
		return true
	case cloudevents.UnsupportedStructuredMode:
		return false
	}

	if deserializer.staticCodec != nil {
		return true
	}

	_, found := deserializer.findSchemaHeader(consumerMsg)
	return found
}

func (deserializer *AvroFileMessageDeserializer) DeserializeKey(_ *sarama.ConsumerMessage) (*DeserializedData, error) {
	return nil, errors.New("AvroFileMessageDeserializer does not support keys")
}

func (deserializer *AvroFileMessageDeserializer) DeserializeValue(consumerMsg *sarama.ConsumerMessage) (*DeserializedData, error) {
	output.Debugf("deserialize value with AvroFileMessageDeserializer")

	if len(consumerMsg.Value) == 0 {
		// tombstone record, can't be unmarshalled as avro
		return &DeserializedData{}, nil
	}

	if contentModeOf(consumerMsg) == cloudevents.StructuredJSONMode {
		return deserializer.deserializeStructuredJSON(consumerMsg)
	}

	codec := deserializer.staticCodec

	if codec == nil {
		headerValue, found := deserializer.findSchemaHeader(consumerMsg)
		if !found {
			return nil, &DeserializationError{errors.Errorf("message has no %q header to resolve the avro schema from", deserializer.schemaHeader)}
		}

		var err error
		codec, err = deserializer.resolveFromMessage(string(headerValue), fmt.Sprintf("%q header", deserializer.schemaHeader))
		if err != nil {
			return nil, err
		}
	}

	textual, err := decodeAvro(codec, consumerMsg.Value)
	if err != nil {
		return nil, err
	}

	return &DeserializedData{schema: codec.Schema(), data: textual}, nil
}

// deserializeStructuredJSON replaces the avro "data_base64" of a JSON format CloudEvent with the decoded "data".
// Events with JSON "data" or without a schema are returned unchanged.
func (deserializer *AvroFileMessageDeserializer) deserializeStructuredJSON(consumerMsg *sarama.ConsumerMessage) (*DeserializedData, error) {
	value := consumerMsg.Value

	event, err := cloudevents.ParseJSONEvent(value)
	if err != nil {
		return nil, &DeserializationError{errors.Wrap(err, "failed to parse structured mode cloudevent")}
	}

	encodedData, found := event.Get("data_base64")
	if !found {
		return &DeserializedData{data: value}, nil
	}

	if _, found := event.Get("data"); found {
		return nil, &DeserializationError{errors.New("structured mode cloudevent must not contain both \"data\" and \"data_base64\"")}
	}

	codec := deserializer.staticCodec

	if codec == nil {
		if rawSchemaURI, found := event.Get("dataschema"); found {
			var schemaURI string
			if err := json.Unmarshal(rawSchemaURI, &schemaURI); err != nil {
				return nil, &DeserializationError{errors.Wrap(err, "cloudevent \"dataschema\" attribute is not a string")}
			}

			if codec, err = deserializer.resolveFromMessage(schemaURI, "cloudevent \"dataschema\" attribute"); err != nil {
				return nil, err
			}
		} else if headerValue, found := deserializer.findSchemaHeader(consumerMsg); found {
			if codec, err = deserializer.resolveFromMessage(string(headerValue), fmt.Sprintf("%q header", deserializer.schemaHeader)); err != nil {
				return nil, err
			}
		} else {
			return &DeserializedData{data: value}, nil
		}
	}

	var base64Data string
	if err := json.Unmarshal(encodedData, &base64Data); err != nil {
		return nil, &DeserializationError{errors.Wrap(err, "cloudevent \"data_base64\" member is not a string")}
	}

	avroData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, &DeserializationError{errors.Wrap(err, "failed to decode cloudevent \"data_base64\" member")}
	}

	textual, err := decodeAvro(codec, avroData)
	if err != nil {
		return nil, err
	}

	event.Replace("data_base64", "data", textual)

	return &DeserializedData{schema: codec.Schema(), data: event.Marshal()}, nil
}

// contentModeOf bridges kafka headers to the cloudevents binding, which takes the content-type value.
func contentModeOf(consumerMsg *sarama.ConsumerMessage) cloudevents.ContentMode {
	contentType, _ := findHeader(consumerMsg.Headers, cloudevents.ContentTypeHeader)
	return cloudevents.ContentModeOf(string(contentType))
}

// findSchemaHeader looks up the configured --avro-schema-header. Without one configured there is
// nothing to look up: searching for the empty name would match a header with an empty key.
func (deserializer *AvroFileMessageDeserializer) findSchemaHeader(consumerMsg *sarama.ConsumerMessage) ([]byte, bool) {
	if deserializer.schemaHeader == "" {
		return nil, false
	}
	return findHeader(consumerMsg.Headers, deserializer.schemaHeader)
}

// resolveFromMessage resolves a schema location taken from message data (a header or a cloudevent attribute).
func (deserializer *AvroFileMessageDeserializer) resolveFromMessage(location, origin string) (*goavro.Codec, error) {
	if location == "" {
		return nil, &DeserializationError{errors.Errorf("%s is present but empty, can't resolve the avro schema", origin)}
	}
	if !avro.IsSchemaURI(location) {
		// plain paths in message data could read arbitrary relative files
		return nil, &DeserializationError{errors.Errorf("%s value %q is not an http(s) or file uri", origin, location)}
	}

	codec, err := deserializer.source.Resolve(location)
	if err != nil {
		// the location came from the message, so a broken one is bad data rather than bad config
		return nil, &DeserializationError{errors.Wrap(err, origin)}
	}
	return codec, nil
}

func decodeAvro(codec *goavro.Codec, data []byte) ([]byte, error) {
	native, _, err := codec.NativeFromBinary(data)
	if err != nil {
		return nil, &DeserializationError{errors.Wrap(err, "failed to parse avro data")}
	}

	textual, err := codec.TextualFromNative(nil, native)
	if err != nil {
		return nil, &DeserializationError{errors.Wrap(err, "failed to convert avro data to json")}
	}

	return textual, nil
}
