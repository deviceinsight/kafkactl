package consume

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"strings"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

const (
	cloudEventsContentTypeHeader         = "content-type"
	cloudEventsStructuredMediaTypePrefix = "application/cloudevents"
	cloudEventsJSONMediaType             = "application/cloudevents+json"
)

// contentMode as defined by the CloudEvents Kafka protocol binding
type contentMode int

const (
	binaryContentMode contentMode = iota
	structuredJSONContentMode
	unsupportedStructuredContentMode
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

	switch messageContentMode(consumerMsg) {
	case structuredJSONContentMode:
		return true
	case unsupportedStructuredContentMode:
		return false
	}

	if deserializer.staticCodec != nil {
		return true
	}

	if deserializer.schemaHeader == "" {
		return false
	}

	_, found := findHeader(consumerMsg.Headers, deserializer.schemaHeader)
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

	if messageContentMode(consumerMsg) == structuredJSONContentMode {
		return deserializer.deserializeStructuredJSON(consumerMsg)
	}

	codec := deserializer.staticCodec

	if codec == nil {
		headerValue, found := findHeader(consumerMsg.Headers, deserializer.schemaHeader)
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

	event, err := parseJSONEvent(value)
	if err != nil {
		return nil, &DeserializationError{errors.Wrap(err, "failed to parse structured mode cloudevent")}
	}

	encodedData, found := event.get("data_base64")
	if !found {
		return &DeserializedData{data: value}, nil
	}

	if _, found := event.get("data"); found {
		return nil, &DeserializationError{errors.New("structured mode cloudevent must not contain both \"data\" and \"data_base64\"")}
	}

	codec := deserializer.staticCodec

	if codec == nil {
		if rawSchemaURI, found := event.get("dataschema"); found {
			var schemaURI string
			if err := json.Unmarshal(rawSchemaURI, &schemaURI); err != nil {
				return nil, &DeserializationError{errors.Wrap(err, "cloudevent \"dataschema\" attribute is not a string")}
			}

			if codec, err = deserializer.resolveFromMessage(schemaURI, "cloudevent \"dataschema\" attribute"); err != nil {
				return nil, err
			}
		} else if headerValue, found := findHeader(consumerMsg.Headers, deserializer.schemaHeader); found && deserializer.schemaHeader != "" {
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

	event.replace("data_base64", "data", textual)

	return &DeserializedData{schema: codec.Schema(), data: event.marshal()}, nil
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

// messageContentMode: a content-type prefixed with "application/cloudevents" means structured mode, else binary mode
func messageContentMode(consumerMsg *sarama.ConsumerMessage) contentMode {
	contentType, found := findHeader(consumerMsg.Headers, cloudEventsContentTypeHeader)
	if !found {
		return binaryContentMode
	}

	mediaType, _, err := mime.ParseMediaType(string(contentType))
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(string(contentType)))
	}

	switch {
	case mediaType == cloudEventsJSONMediaType:
		return structuredJSONContentMode
	case strings.HasPrefix(mediaType, cloudEventsStructuredMediaTypePrefix):
		return unsupportedStructuredContentMode
	default:
		return binaryContentMode
	}
}

func findHeader(headers []*sarama.RecordHeader, name string) ([]byte, bool) {
	for _, header := range headers {
		if header.Key != nil && string(header.Key) == name {
			return decodeAMQPValue(header.Value), true
		}
	}
	return nil, false
}
