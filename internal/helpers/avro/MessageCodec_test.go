package avro

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const testRecordSchema = `{
  "type": "record",
  "name": "Message",
  "fields": [
    {"name": "text", "type": "string"},
    {"name": "count", "type": "long"}
  ]
}`

const testUnionSchema = `{
  "type": "record",
  "name": "UnionMessage",
  "fields": [
    {"name": "value", "type": ["null", "string"]}
  ]
}`

func TestMessageCodecRoundTrip(t *testing.T) {
	textual := []byte(`{"text":"hello","count":3}`)

	for _, test := range []struct {
		name      string
		jsonCodec JSONCodec
	}{
		{name: "standard JSON", jsonCodec: Standard},
		{name: "Avro JSON", jsonCodec: Avro},
	} {
		t.Run(test.name, func(t *testing.T) {
			codec, err := NewMessageCodec(testRecordSchema, test.jsonCodec)
			if err != nil {
				t.Fatalf("NewMessageCodec(): unexpected error: %v", err)
			}
			if codec.Schema() != testRecordSchema {
				t.Fatalf("Schema() = %q, want %q", codec.Schema(), testRecordSchema)
			}

			binary, err := codec.EncodeTextual(textual)
			if err != nil {
				t.Fatalf("EncodeTextual(): unexpected error: %v", err)
			}
			if len(binary) == 0 {
				t.Fatal("EncodeTextual() returned an empty Avro payload")
			}

			decoded, err := codec.DecodeBinary(binary)
			if err != nil {
				t.Fatalf("DecodeBinary(): unexpected error: %v", err)
			}
			assertJSONEqual(t, decoded, textual)
		})
	}
}

func TestMessageCodecUsesConfiguredJSONRepresentation(t *testing.T) {
	for _, test := range []struct {
		name      string
		jsonCodec JSONCodec
		textual   []byte
	}{
		{name: "standard JSON", jsonCodec: Standard, textual: []byte(`{"value":"hello"}`)},
		{name: "Avro JSON", jsonCodec: Avro, textual: []byte(`{"value":{"string":"hello"}}`)},
	} {
		t.Run(test.name, func(t *testing.T) {
			codec, err := NewMessageCodec(testUnionSchema, test.jsonCodec)
			if err != nil {
				t.Fatalf("NewMessageCodec(): unexpected error: %v", err)
			}

			binary, err := codec.EncodeTextual(test.textual)
			if err != nil {
				t.Fatalf("EncodeTextual(): unexpected error: %v", err)
			}
			decoded, err := codec.DecodeBinary(binary)
			if err != nil {
				t.Fatalf("DecodeBinary(): unexpected error: %v", err)
			}
			assertJSONEqual(t, decoded, test.textual)
		})
	}
}

func TestMessageCodecAcceptsJSONWhitespace(t *testing.T) {
	codec, err := NewMessageCodec(`"long"`, Standard)
	if err != nil {
		t.Fatalf("NewMessageCodec(): unexpected error: %v", err)
	}

	binary, err := codec.EncodeTextual([]byte(" \t\r\n42 \t\r\n"))
	if err != nil {
		t.Fatalf("EncodeTextual(): unexpected error: %v", err)
	}
	decoded, err := codec.DecodeBinary(binary)
	if err != nil {
		t.Fatalf("DecodeBinary(): unexpected error: %v", err)
	}
	assertJSONEqual(t, decoded, []byte("42"))
}

func TestMessageCodecRoundTripsZeroByteValue(t *testing.T) {
	codec, err := NewMessageCodec(`"null"`, Standard)
	if err != nil {
		t.Fatalf("NewMessageCodec(): unexpected error: %v", err)
	}

	binary, err := codec.EncodeTextual([]byte("null"))
	if err != nil {
		t.Fatalf("EncodeTextual(): unexpected error: %v", err)
	}
	if binary == nil {
		t.Fatal("EncodeTextual() returned a nil Avro payload for a zero-byte value")
	}
	if len(binary) != 0 {
		t.Fatalf("EncodeTextual() returned %d bytes, want 0", len(binary))
	}

	decoded, err := codec.DecodeBinary(binary)
	if err != nil {
		t.Fatalf("DecodeBinary(): unexpected error: %v", err)
	}
	assertJSONEqual(t, decoded, []byte("null"))
}

func TestNewMessageCodecRejectsEmptyAndInvalidSchemas(t *testing.T) {
	for _, test := range []struct {
		name        string
		schema      string
		wantErrPart string
	}{
		{name: "empty", schema: "", wantErrPart: "must not be empty"},
		{name: "whitespace", schema: " \r\n\t", wantErrPart: "must not be empty"},
		{name: "invalid", schema: `{not-json`, wantErrPart: "failed to parse avro schema"},
	} {
		t.Run(test.name, func(t *testing.T) {
			codec, err := NewMessageCodec(test.schema, Standard)
			if err == nil {
				t.Fatalf(
					"expected NewMessageCodec() to return an error, got nil with codec %#v",
					codec,
				)
			}
			if !strings.Contains(err.Error(), test.wantErrPart) {
				t.Fatalf("NewMessageCodec() error = %q, want it to contain %q", err, test.wantErrPart)
			}
		})
	}
}

func TestNewMessageCodecRejectsUnsupportedJSONCodec(t *testing.T) {
	codec, err := NewMessageCodec(testRecordSchema, JSONCodec(99))
	if err == nil {
		t.Fatalf(
			"expected NewMessageCodec() to return an error, got nil with codec %#v",
			codec,
		)
	}
	const wantErrPart = "unsupported avro JSON codec: 99"
	if !strings.Contains(err.Error(), wantErrPart) {
		t.Fatalf("NewMessageCodec() error = %q, want it to contain %q", err, wantErrPart)
	}
}

func TestMessageCodecReportsConversionErrors(t *testing.T) {
	codec, err := NewMessageCodec(testRecordSchema, Standard)
	if err != nil {
		t.Fatalf("NewMessageCodec(): unexpected error: %v", err)
	}

	const encodeErrPart = "failed to convert textual data to avro data"
	if _, err = codec.EncodeTextual([]byte(`{"text":42,"count":"bad"}`)); err == nil {
		t.Fatal("EncodeTextual() returned nil error for invalid textual data")
	} else if !strings.Contains(err.Error(), encodeErrPart) {
		t.Fatalf("EncodeTextual() error = %q, want it to contain %q", err, encodeErrPart)
	} else if err.Error() == encodeErrPart {
		t.Fatalf("EncodeTextual() error = %q, want the underlying goavro error details", err)
	}

	const decodeErrPart = "failed to decode avro data"
	if _, err = codec.DecodeBinary([]byte{0xff}); err == nil {
		t.Fatal("DecodeBinary() returned nil error for invalid binary data")
	} else if !strings.Contains(err.Error(), decodeErrPart) {
		t.Fatalf("DecodeBinary() error = %q, want it to contain %q", err, decodeErrPart)
	} else if err.Error() == decodeErrPart {
		t.Fatalf("DecodeBinary() error = %q, want the underlying goavro error details", err)
	}
}

func TestMessageCodecRejectsInvalidJSON(t *testing.T) {
	for _, test := range []struct {
		name    string
		schema  string
		textual []byte
	}{
		{
			name:    "non-JSON trailing data",
			schema:  testRecordSchema,
			textual: []byte(`{"text":"hello","count":3} trailing`),
		},
		{
			name:    "second JSON value",
			schema:  testRecordSchema,
			textual: []byte(`{"text":"hello","count":3} {}`),
		},
		{
			name:    "vertical tab",
			schema:  testRecordSchema,
			textual: append([]byte(`{"text":"hello","count":3}`), '\v'),
		},
		{
			name:    "non-breaking space",
			schema:  testRecordSchema,
			textual: append([]byte(`{"text":"hello","count":3}`), []byte("\u00a0")...),
		},
		{
			name:    "invalid escape",
			schema:  `"string"`,
			textual: []byte(`"A\q"`),
		},
		{
			name:    "invalid UTF-8",
			schema:  `"string"`,
			textual: []byte{'"', 0xff, '"'},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			codec, err := NewMessageCodec(test.schema, Standard)
			if err != nil {
				t.Fatalf("NewMessageCodec(): unexpected error: %v", err)
			}

			_, err = codec.EncodeTextual(test.textual)
			if err == nil {
				t.Fatal("EncodeTextual() returned nil error for invalid JSON")
			}
			const wantErrPart = "invalid JSON"
			if !strings.Contains(err.Error(), wantErrPart) {
				t.Fatalf("EncodeTextual() error = %q, want it to contain %q", err, wantErrPart)
			}
		})
	}
}

func TestMessageCodecRejectsTrailingData(t *testing.T) {
	codec, err := NewMessageCodec(`"long"`, Standard)
	if err != nil {
		t.Fatalf("NewMessageCodec(): unexpected error: %v", err)
	}

	const trailingDataErrPart = "unexpected trailing data"
	if _, err = codec.EncodeTextual([]byte("1.5")); err == nil {
		t.Fatal("EncodeTextual() returned nil error for trailing textual data")
	} else if !strings.Contains(err.Error(), trailingDataErrPart) {
		t.Fatalf("EncodeTextual() error = %q, want it to contain %q", err, trailingDataErrPart)
	}

	binary, err := codec.EncodeTextual([]byte("1"))
	if err != nil {
		t.Fatalf("EncodeTextual(): unexpected error: %v", err)
	}
	binary = append(binary, 0)
	if _, err = codec.DecodeBinary(binary); err == nil {
		t.Fatal("DecodeBinary() returned nil error for trailing binary data")
	} else if !strings.Contains(err.Error(), trailingDataErrPart) {
		t.Fatalf("DecodeBinary() error = %q, want it to contain %q", err, trailingDataErrPart)
	}
}

func assertJSONEqual(t *testing.T, actual, expected []byte) {
	t.Helper()

	var actualValue any
	if err := json.Unmarshal(actual, &actualValue); err != nil {
		t.Fatalf("actual value is not valid JSON: %v; value = %q", err, actual)
	}

	var expectedValue any
	if err := json.Unmarshal(expected, &expectedValue); err != nil {
		t.Fatalf("expected value is not valid JSON: %v; value = %q", err, expected)
	}

	if !reflect.DeepEqual(actualValue, expectedValue) {
		t.Fatalf("decoded JSON = %s, want %s", actual, expected)
	}
}
