package consume

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

// failingDeserializer decodes every message as-is except those whose value
// equals "bad", which fail deserialization.
type failingDeserializer struct{}

func (d *failingDeserializer) CanDeserializeKey(*sarama.ConsumerMessage, Flags) bool   { return true }
func (d *failingDeserializer) CanDeserializeValue(*sarama.ConsumerMessage, Flags) bool { return true }

func (d *failingDeserializer) DeserializeKey(msg *sarama.ConsumerMessage) (*DeserializedData, error) {
	return &DeserializedData{data: msg.Key}, nil
}

func (d *failingDeserializer) DeserializeValue(msg *sarama.ConsumerMessage) (*DeserializedData, error) {
	if string(msg.Value) == "bad" {
		return nil, errors.New("boom")
	}
	return &DeserializedData{data: msg.Value}, nil
}

func captureStreams(t *testing.T) (out, errOut *bytes.Buffer) {
	t.Helper()
	out = new(bytes.Buffer)
	errOut = new(bytes.Buffer)
	previous := output.IoStreams
	output.IoStreams = output.IOStreams{Out: out, ErrOut: errOut, DebugOut: io.Discard}
	t.Cleanup(func() { output.IoStreams = previous })
	return out, errOut
}

func drain(t *testing.T, flags Flags, values ...string) (out, errOut *bytes.Buffer, err error) {
	t.Helper()

	out, errOut = captureStreams(t)

	filter, ferr := NewMessageFilter("", "", nil)
	if ferr != nil {
		t.Fatalf("failed to create filter: %v", ferr)
	}

	messages := make(chan *sarama.ConsumerMessage, len(values))
	for i, v := range values {
		messages <- &sarama.ConsumerMessage{Partition: 0, Offset: int64(i), Value: []byte(v)}
	}
	close(messages)

	stopConsumers := make(chan bool, 1)
	chain := MessageDeserializerChain{&failingDeserializer{}}

	group := deserializeMessages(context.Background(), flags, messages, stopConsumers, chain, filter)
	return out, errOut, group.Wait()
}

func TestDeserializeMessages_IgnoreErrorsSkipsBadMessage(t *testing.T) {
	out, errOut, err := drain(t, Flags{IgnoreErrors: true}, "one", "bad", "three")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdout := out.String()
	if !strings.Contains(stdout, "one") || !strings.Contains(stdout, "three") {
		t.Fatalf("expected good messages to be printed, got: %q", stdout)
	}
	if strings.Contains(stdout, "bad") {
		t.Fatalf("did not expect the failing message to be printed, got: %q", stdout)
	}
	if warning := errOut.String(); !strings.Contains(warning, "skipping message on partition 0 at offset 1") {
		t.Fatalf("expected a skip warning for the failing message, got: %q", warning)
	}
}

func TestDeserializeMessages_WithoutIgnoreErrorsAborts(t *testing.T) {
	_, _, err := drain(t, Flags{IgnoreErrors: false}, "one", "bad", "three")
	if err == nil {
		t.Fatal("expected an error to abort consumption")
	}
	if !isDeserializationError(err) {
		t.Fatalf("expected a deserialization error, got: %v", err)
	}
}

func TestIsDeserializationError(t *testing.T) {
	t.Parallel()

	if isDeserializationError(errors.New("plain")) {
		t.Fatal("a plain error must not be classified as a deserialization error")
	}

	inner := errors.New("bad payload")
	deser := &DeserializationError{inner}
	if !isDeserializationError(deser) {
		t.Fatal("a DeserializationError must be classified as one")
	}
	if got := deser.Error(); got != "bad payload" {
		t.Fatalf("unexpected error message: %q", got)
	}
	if !errors.Is(deser, inner) {
		t.Fatal("DeserializationError must unwrap to its cause")
	}
}
