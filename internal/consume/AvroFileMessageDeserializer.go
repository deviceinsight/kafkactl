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

// contentMode is the CloudEvents Kafka protocol binding content mode of a message.
type contentMode int

const (
	// binaryContentMode: the value is the raw avro data (also used for messages that aren't CloudEvents at all).
	binaryContentMode contentMode = iota
	// structuredJSONContentMode: the value is a single CloudEvent in the JSON event format.
	structuredJSONContentMode
	// unsupportedStructuredContentMode: any other event format (batch, protobuf, avro, ...), printed as-is.
	unsupportedStructuredContentMode
)

// AvroFileMessageDeserializer decodes avro messages without a Confluent schema registry,
// resolving the schema either from a static file/url (--avro-schema-file) or, per message,
// from a header (--avro-schema-header, e.g. the CloudEvents "ce_dataschema" header) whose
// value is a URI pointing at the schema.
//
// CloudEvents in structured content mode with the JSON event format are supported as well:
// their base64 encoded "data_base64" member is decoded with the static schema, the schema
// referenced by the event's "dataschema" attribute or the one in the schema header.
type AvroFileMessageDeserializer struct {
	schemaHeader string
	staticCodec  *goavro.Codec
	source       *avro.SchemaSource
}

// CreateAvroFileMessageDeserializer resolves schemaFile once, if set, so that a static
// --avro-schema-file schema isn't re-fetched/re-compiled on every consumed message. When
// schemaFile is empty, the schema is instead resolved per message from schemaHeader.
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
			return nil, errors.Errorf("message has no %q header to resolve the avro schema from", deserializer.schemaHeader)
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

// deserializeStructuredJSON decodes the "data_base64" member of a CloudEvent in the JSON event format and replaces
// it with a "data" member holding the decoded avro data. The schema is the static one, else the one referenced by the
// event's "dataschema" attribute, else the one referenced by the schema header. Events with JSON "data", or without a
// schema to decode "data_base64" with, are returned unchanged.
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
			// structured mode messages may carry the same headers as binary mode ones
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
		// only absolute uris are accepted from message data, so that a plain path in a message
		// can't make kafkactl read files relative to the working directory
		return nil, &DeserializationError{errors.Errorf("%s value %q is not an http(s) or file uri", origin, location)}
	}
	return deserializer.source.Resolve(location)
}

func decodeAvro(codec *goavro.Codec, data []byte) ([]byte, error) {
	native, _, err := codec.NativeFromBinary(data)
	if err != nil {
		return nil, &DeserializationError{errors.Wrap(err, "failed to parse avro data")}
	}

	textual, err := codec.TextualFromNative(nil, native)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert data to avro data")
	}

	return textual, nil
}

// messageContentMode determines the content mode as defined by the CloudEvents Kafka protocol binding: a content-type
// header prefixed with "application/cloudevents" denotes structured content mode, anything else binary content mode.
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
