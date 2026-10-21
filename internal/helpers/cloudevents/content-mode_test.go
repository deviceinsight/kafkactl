package cloudevents

import "testing"

func TestContentModeOf(t *testing.T) {
	t.Parallel()

	for contentType, expected := range map[string]ContentMode{
		"":                             BinaryMode,
		"application/avro":             BinaryMode,
		"application/json":             BinaryMode,
		"application/cloudevents+json": StructuredJSONMode,
		"application/cloudevents+json; charset=UTF-8": StructuredJSONMode,
		// the binding matches media types case insensitively
		"Application/CloudEvents+JSON":       StructuredJSONMode,
		"  application/cloudevents+json  ":   StructuredJSONMode,
		"application/cloudevents-batch+json": UnsupportedStructuredMode,
		"application/cloudevents+protobuf":   UnsupportedStructuredMode,
		"application/cloudevents+avro":       UnsupportedStructuredMode,
		// only the parameters are malformed, the media type still classifies
		"application/cloudevents+json; =bad":     StructuredJSONMode,
		"application/cloudevents+json; charset=": StructuredJSONMode,
		// the media type itself is unparseable, so nothing can be claimed about it
		"application/cloudevents+json/extra": BinaryMode,
		"application/":                       BinaryMode,
		"/":                                  BinaryMode,
	} {
		if got := ContentModeOf(contentType); got != expected {
			t.Errorf("ContentModeOf(%q) = %v, expected %v", contentType, got, expected)
		}
	}
}
