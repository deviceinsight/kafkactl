package cloudevents

import (
	"testing"
)

// ParseJSONEvent keeps member values as json.RawMessage rather than decoding them, which is what
// lets an event round trip without its numbers being reformatted or losing precision.
func TestParseJSONEvent_PreservesNumbersAndOrder(t *testing.T) {
	t.Parallel()

	event := `{"big":12345678901234567890,"float":1.0,"exp":1e400,"negzero":-0.0,` +
		`"precise":0.10000000000000000001,"z":"last","a":"first"}`

	parsed, err := ParseJSONEvent([]byte(event))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := string(parsed.Marshal()); got != event {
		t.Fatalf("expected the event to round trip verbatim\n in: %s\nout: %s", event, got)
	}
}

func TestParseJSONEvent_Malformed(t *testing.T) {
	t.Parallel()

	for name, event := range map[string]string{
		"empty":                 ``,
		"truncated object":      `{`,
		"truncated member":      `{"a"`,
		"mismatched delimiter":  `{]`,
		"trailing comma":        `{"a":1,}`,
		"array":                 `[1]`,
		"data after the object": `{"a":1}x`,
	} {
		// must report rather than panic: the member name used to be an unchecked type assertion
		if _, err := ParseJSONEvent([]byte(event)); err == nil {
			t.Errorf("%s: expected an error, got nil", name)
		}
	}
}
