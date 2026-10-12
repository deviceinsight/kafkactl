package consume

import (
	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

// AvroFileMessageDeserializer decodes avro messages without a Confluent schema registry,
// resolving the schema either from a static file/url (--avro-schema-file) or, per message,
// from a header (--avro-schema-header, e.g. the CloudEvents "ce_dataschema" header) whose
// value is a URI pointing at the schema.
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

	codec := deserializer.staticCodec

	if codec == nil {
		headerValue, found := findHeader(consumerMsg.Headers, deserializer.schemaHeader)
		if !found {
			return nil, errors.Errorf("message has no %q header to resolve the avro schema from", deserializer.schemaHeader)
		}

		location := string(headerValue)
		if location == "" {
			return nil, &DeserializationError{errors.Errorf("%q header is present but empty, can't resolve the avro schema", deserializer.schemaHeader)}
		}
		if !avro.IsSchemaURI(location) {
			// only absolute uris are accepted from message data, so that a plain path in a header
			// can't make kafkactl read files relative to the working directory
			return nil, &DeserializationError{errors.Errorf("%q header value %q is not an http(s) or file uri", deserializer.schemaHeader, location)}
		}

		var err error
		codec, err = deserializer.source.Resolve(location)
		if err != nil {
			return nil, err
		}
	}

	native, _, err := codec.NativeFromBinary(consumerMsg.Value)
	if err != nil {
		return nil, &DeserializationError{errors.Wrap(err, "failed to parse avro data")}
	}

	textual, err := codec.TextualFromNative(nil, native)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert data to avro data")
	}

	return &DeserializedData{schema: codec.Schema(), data: textual}, nil
}

func findHeader(headers []*sarama.RecordHeader, name string) ([]byte, bool) {
	for _, header := range headers {
		if header.Key != nil && string(header.Key) == name {
			return decodeAMQPValue(header.Value), true
		}
	}
	return nil, false
}
