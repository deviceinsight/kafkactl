package consume

import (
	"strings"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/helpers/avro"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
)

// RawAvroMessageDeserializer decodes unframed Avro message values with a
// schema loaded from a static source or referenced by a Kafka record header.
type RawAvroMessageDeserializer struct {
	schemaHeader string
	staticCodec  *avro.MessageCodec
	resolver     *avro.SchemaResolver
}

func NewRawAvroMessageDeserializer(schemaHeader string, staticCodec *avro.MessageCodec,
	resolver *avro.SchemaResolver,
) *RawAvroMessageDeserializer {
	return &RawAvroMessageDeserializer{
		schemaHeader: schemaHeader,
		staticCodec:  staticCodec,
		resolver:     resolver,
	}
}

func (deserializer *RawAvroMessageDeserializer) CanDeserializeKey(_ *sarama.ConsumerMessage, _ Flags) bool {
	return false
}

func (deserializer *RawAvroMessageDeserializer) CanDeserializeValue(msg *sarama.ConsumerMessage, _ Flags) bool {
	// Kafka tombstones have a nil value and must remain nil. An empty, non-nil
	// byte slice can still be a valid Avro payload for some schemas.
	return msg.Value != nil
}

func (deserializer *RawAvroMessageDeserializer) DeserializeKey(_ *sarama.ConsumerMessage) (*DeserializedData, error) {
	return nil, errors.New("raw avro deserialization is only supported for message values")
}

func (deserializer *RawAvroMessageDeserializer) DeserializeValue(msg *sarama.ConsumerMessage) (*DeserializedData, error) {
	codec, err := deserializer.resolveCodec(msg)
	if err != nil {
		return nil, err
	}

	output.Debugf("deserialize value with RawAvroMessageDeserializer")
	data, err := codec.DecodeBinary(msg.Value)
	if err != nil {
		return nil, err
	}

	return &DeserializedData{schema: codec.Schema(), data: data}, nil
}

func (deserializer *RawAvroMessageDeserializer) resolveCodec(msg *sarama.ConsumerMessage) (*avro.MessageCodec, error) {
	if deserializer.schemaHeader != "" {
		if schemaURI, found := findLastHeaderValue(msg.Headers, deserializer.schemaHeader); found {
			schemaURI = strings.TrimSpace(schemaURI)
			if schemaURI == "" {
				return nil, errors.Errorf("avro schema header %q is empty", deserializer.schemaHeader)
			}
			if deserializer.resolver == nil {
				return nil, errors.New("avro schema URI resolver is not configured")
			}

			codec, err := deserializer.resolver.ResolveURI(schemaURI)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to resolve avro schema from header %q", deserializer.schemaHeader)
			}
			return codec, nil
		}
	}

	if deserializer.staticCodec != nil {
		return deserializer.staticCodec, nil
	}

	if deserializer.schemaHeader != "" {
		return nil, errors.Errorf("avro schema header %q is missing", deserializer.schemaHeader)
	}

	return nil, errors.New("no avro schema source is configured")
}

func findLastHeaderValue(headers []*sarama.RecordHeader, name string) (string, bool) {
	for i := len(headers) - 1; i >= 0; i-- {
		header := headers[i]
		if header == nil || header.Key == nil || string(header.Key) != name {
			continue
		}

		return string(decodeAMQPValue(header.Value)), true
	}

	return "", false
}
