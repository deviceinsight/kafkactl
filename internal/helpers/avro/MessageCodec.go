package avro

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

const jsonWhitespace = " \t\r\n"

// MessageCodec converts between textual JSON and unframed Avro binary data.
type MessageCodec struct {
	schema string
	codec  *goavro.Codec
}

// NewMessageCodec compiles schema for the configured Avro JSON representation.
func NewMessageCodec(schema string, jsonCodec JSONCodec) (*MessageCodec, error) {
	if strings.TrimSpace(schema) == "" {
		return nil, errors.New("avro schema must not be empty")
	}

	var (
		codec *goavro.Codec
		err   error
	)

	switch jsonCodec {
	case Standard:
		codec, err = goavro.NewCodecForStandardJSONFull(schema)
	case Avro:
		codec, err = goavro.NewCodec(schema)
	default:
		return nil, errors.Errorf("unsupported avro JSON codec: %d", jsonCodec)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse avro schema")
	}

	return &MessageCodec{schema: schema, codec: codec}, nil
}

// Schema returns the schema used to compile the codec.
func (codec *MessageCodec) Schema() string {
	return codec.schema
}

// EncodeTextual converts textual JSON to unframed Avro binary data.
func (codec *MessageCodec) EncodeTextual(data []byte) ([]byte, error) {
	native, remainder, err := codec.codec.NativeFromTextual(bytes.TrimLeft(data, jsonWhitespace))
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert textual data to avro data")
	}
	// Keep goavro's detailed conversion errors, but reject non-standard forms
	// that its textual decoders may otherwise accept.
	if !utf8.Valid(data) || !json.Valid(data) {
		return nil, errors.New("failed to convert textual data to avro data: invalid JSON")
	}
	if len(bytes.Trim(remainder, jsonWhitespace)) != 0 {
		return nil, errors.New("failed to convert textual data to avro data: unexpected trailing data")
	}

	// Some valid Avro values, such as an empty record or null, have a zero-byte
	// wire representation. Start with a non-nil slice so those values remain
	// distinguishable from a Kafka tombstone.
	binary, err := codec.codec.BinaryFromNative(make([]byte, 0), native)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode avro data")
	}

	return binary, nil
}

// DecodeBinary converts unframed Avro binary data to textual JSON.
func (codec *MessageCodec) DecodeBinary(data []byte) ([]byte, error) {
	native, remainder, err := codec.codec.NativeFromBinary(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode avro data")
	}
	if len(remainder) != 0 {
		return nil, errors.New("failed to decode avro data: unexpected trailing data")
	}

	textual, err := codec.codec.TextualFromNative(nil, native)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert avro data to textual data")
	}

	return textual, nil
}
