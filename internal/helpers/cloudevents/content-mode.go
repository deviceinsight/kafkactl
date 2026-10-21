// Package cloudevents implements the parts of the CloudEvents spec that kafkactl needs to decode
// event payloads. It deliberately knows nothing about Kafka: callers extract header values and
// pass them in.
package cloudevents

import (
	"mime"
	"strings"
)

// ContentTypeHeader carries the content mode, see ContentModeOf.
const ContentTypeHeader = "content-type"

const (
	structuredMediaTypePrefix = "application/cloudevents"
	jsonMediaType             = "application/cloudevents+json"
)

// ContentMode as defined by the CloudEvents Kafka protocol binding.
type ContentMode int

const (
	// BinaryMode has the event attributes in headers and the raw data as the message value.
	BinaryMode ContentMode = iota
	// StructuredJSONMode has the whole event as a JSON object in the message value.
	StructuredJSONMode
	// UnsupportedStructuredMode is a structured event in a format kafkactl cannot decode,
	// such as cloudevents-batch+json, +protobuf or +avro.
	UnsupportedStructuredMode
)

// ContentModeOf reports the content mode a content-type implies: the Kafka protocol binding
// defines a type prefixed with "application/cloudevents" as structured mode, anything else,
// including a missing content-type, as binary mode.
func ContentModeOf(contentType string) ContentMode {
	if contentType == "" {
		return BinaryMode
	}

	// ParseMediaType still returns the lowercased media type when only the parameters are
	// malformed, which is the case worth classifying. It yields "" when the type itself is.
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil && mediaType == "" {
		return BinaryMode
	}

	switch {
	case mediaType == jsonMediaType:
		return StructuredJSONMode
	case strings.HasPrefix(mediaType, structuredMediaTypePrefix):
		return UnsupportedStructuredMode
	default:
		return BinaryMode
	}
}
