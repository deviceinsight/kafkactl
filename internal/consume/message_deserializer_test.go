package consume

import (
	"bytes"
	"testing"

	"regexp"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

type fakeDeserializer struct {
	keyBytes   []byte
	valueBytes []byte
}

func (f *fakeDeserializer) CanDeserializeKey(msg *sarama.ConsumerMessage, flags Flags) bool {
	return f.keyBytes != nil
}
func (f *fakeDeserializer) CanDeserializeValue(msg *sarama.ConsumerMessage, flags Flags) bool {
	return f.valueBytes != nil
}
func (f *fakeDeserializer) DeserializeKey(msg *sarama.ConsumerMessage) (*DeserializedData, error) {
	return &DeserializedData{data: f.keyBytes}, nil
}
func (f *fakeDeserializer) DeserializeValue(msg *sarama.ConsumerMessage) (*DeserializedData, error) {
	return &DeserializedData{data: f.valueBytes}, nil
}

func TestSubstringMatchesValue(t *testing.T) {
	t.Parallel()
	_ = output.NewTestIOStreams(nil)

	// reset output buffer
	buf := output.IoStreams.Out.(*bytes.Buffer)
	buf.Reset()

	flags := Flags{FilterKeyword: "needle", FilterByRegex: false}

	chain := MessageDeserializerChain{&fakeDeserializer{keyBytes: nil, valueBytes: []byte("this contains needle inside")}}

	msg := &sarama.ConsumerMessage{Partition: 0, Offset: 1}

	if err := chain.Deserialize(msg, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out == "" || !containsSubstring(out, "needle") {
		t.Fatalf("expected output to contain keyword; got: %q", out)
	}
}

func TestSubstringNoMatch(t *testing.T) {
	t.Parallel()
	_ = output.NewTestIOStreams(nil)

	buf := output.IoStreams.Out.(*bytes.Buffer)
	buf.Reset()

	flags := Flags{FilterKeyword: "nomatch", FilterByRegex: false}

	chain := MessageDeserializerChain{&fakeDeserializer{keyBytes: nil, valueBytes: []byte("some other content")}}
	msg := &sarama.ConsumerMessage{Partition: 0, Offset: 2}

	if err := chain.Deserialize(msg, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out != "" {
		t.Fatalf("expected no output for non-matching message; got: %q", out)
	}
}

func TestRegexMatchesKey(t *testing.T) {
	t.Parallel()
	_ = output.NewTestIOStreams(nil)

	buf := output.IoStreams.Out.(*bytes.Buffer)
	buf.Reset()

	// Go's regexp (RE2) doesn't support \d shorthand; use [0-9] instead.
	flags := Flags{FilterKeyword: "^k-[0-9]+$", FilterByRegex: true, filterRegexp: regexp.MustCompile("^k-[0-9]+$")}

	chain := MessageDeserializerChain{&fakeDeserializer{keyBytes: []byte("k-1234"), valueBytes: []byte("value")}}
	msg := &sarama.ConsumerMessage{Partition: 1, Offset: 3}
	// sanity check: ensure compiled regex matches the key we're going to send
	if flags.filterRegexp == nil || !flags.filterRegexp.MatchString("k-1234") {
		t.Fatalf("test setup invalid: compiled regex doesn't match key")
	}
	if err := chain.Deserialize(msg, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out == "" || !containsSubstring(out, "value") {
		t.Fatalf("expected output for matching key; got: %q", out)
	}
}

func TestBinaryValueIgnored(t *testing.T) {
	t.Parallel()
	_ = output.NewTestIOStreams(nil)

	buf := output.IoStreams.Out.(*bytes.Buffer)
	buf.Reset()

	flags := Flags{FilterKeyword: "\u0000\u0001", FilterByRegex: false}

	// binary (non-UTF8)
	chain := MessageDeserializerChain{&fakeDeserializer{keyBytes: nil, valueBytes: []byte{0xff, 0xfe, 0xfd}}}
	msg := &sarama.ConsumerMessage{Partition: 1, Offset: 4}

	if err := chain.Deserialize(msg, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out != "" {
		t.Fatalf("expected no output for binary non-textual value; got: %q", out)
	}
}

func TestFilterModeKeyOnly(t *testing.T) {
	t.Parallel()
	_ = output.NewTestIOStreams(nil)

	buf := output.IoStreams.Out.(*bytes.Buffer)
	buf.Reset()

	// key contains keyword, value does not
	flags := Flags{FilterKeyword: "needle", FilterByRegex: false, FilterMode: string(FilterModeKey)}

	chain := MessageDeserializerChain{&fakeDeserializer{keyBytes: []byte("has needle in key"), valueBytes: []byte("value without")}}
	msg := &sarama.ConsumerMessage{Partition: 2, Offset: 10}

	if err := chain.Deserialize(msg, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatalf("expected output for key-only matching mode; got empty output")
	}
}

func TestFilterModeKeyOnlyIgnoresValue(t *testing.T) {
	t.Parallel()
	_ = output.NewTestIOStreams(nil)

	buf := output.IoStreams.Out.(*bytes.Buffer)
	buf.Reset()

	// value contains keyword, key does not -> should NOT match in key-only mode
	flags := Flags{FilterKeyword: "needle", FilterByRegex: false, FilterMode: string(FilterModeKey)}

	chain := MessageDeserializerChain{&fakeDeserializer{keyBytes: []byte("no match"), valueBytes: []byte("contains needle")}}
	msg := &sarama.ConsumerMessage{Partition: 3, Offset: 11}

	if err := chain.Deserialize(msg, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out != "" {
		t.Fatalf("expected no output for key-only mode when only value matches; got: %q", out)
	}
}

func TestFilterModeMessageOnly(t *testing.T) {
	t.Parallel()
	_ = output.NewTestIOStreams(nil)

	buf := output.IoStreams.Out.(*bytes.Buffer)
	buf.Reset()

	// value contains keyword, key does not
	flags := Flags{FilterKeyword: "needle", FilterByRegex: false, FilterMode: string(FilterModeMessage)}

	chain := MessageDeserializerChain{&fakeDeserializer{keyBytes: []byte("nope"), valueBytes: []byte("value has needle")}}
	msg := &sarama.ConsumerMessage{Partition: 4, Offset: 12}

	if err := chain.Deserialize(msg, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatalf("expected output for message-only matching mode; got empty output")
	}
}

func TestFilterModeAnyMatchesEither(t *testing.T) {
	t.Parallel()
	_ = output.NewTestIOStreams(nil)

	buf := output.IoStreams.Out.(*bytes.Buffer)
	buf.Reset()

	// key doesn't match but value does -> any should match
	flags := Flags{FilterKeyword: "needle", FilterByRegex: false, FilterMode: string(FilterModeAny)}

	chain := MessageDeserializerChain{&fakeDeserializer{keyBytes: []byte("nope"), valueBytes: []byte("value contains needle")}}
	msg := &sarama.ConsumerMessage{Partition: 5, Offset: 13}

	if err := chain.Deserialize(msg, flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Fatalf("expected output for FilterModeAny when value matches; got empty output")
	}
}

// small helper copied to avoid bringing extra deps
func containsSubstring(s, sub string) bool { return bytes.Contains([]byte(s), []byte(sub)) }
