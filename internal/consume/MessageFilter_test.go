package consume

import (
	"testing"

	"github.com/IBM/sarama"
)

func TestMessageFilter_GlobMatchesKey(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("user-*", "", nil)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	key := &DeserializedData{data: []byte("user-123")}
	value := &DeserializedData{data: []byte("some value")}
	msg := &sarama.ConsumerMessage{}

	if !filter.Matches(msg, key, value) {
		t.Fatal("expected key 'user-123' to match pattern 'user-*'")
	}
}

func TestMessageFilter_GlobNoMatchKey(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("user-*", "", nil)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	key := &DeserializedData{data: []byte("admin-456")}
	value := &DeserializedData{data: []byte("some value")}
	msg := &sarama.ConsumerMessage{}

	if filter.Matches(msg, key, value) {
		t.Fatal("expected key 'admin-456' to NOT match pattern 'user-*'")
	}
}

func TestMessageFilter_GlobMatchesValue(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("", "*error*", nil)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	key := &DeserializedData{data: []byte("key")}
	value := &DeserializedData{data: []byte("this is an error message")}
	msg := &sarama.ConsumerMessage{}

	if !filter.Matches(msg, key, value) {
		t.Fatal("expected value to match pattern '*error*'")
	}
}

func TestMessageFilter_GlobAlternatives(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("{user,admin}-*", "", nil)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	testCases := []struct {
		key      string
		expected bool
	}{
		{"user-123", true},
		{"admin-456", true},
		{"guest-789", false},
	}

	for _, tc := range testCases {
		key := &DeserializedData{data: []byte(tc.key)}
		value := &DeserializedData{data: []byte("value")}
		msg := &sarama.ConsumerMessage{}

		matched := filter.Matches(msg, key, value)
		if matched != tc.expected {
			t.Errorf("key %q: expected match=%v, got %v", tc.key, tc.expected, matched)
		}
	}
}

func TestMessageFilter_BinaryValueIgnored(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("", "test", nil)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	key := &DeserializedData{data: []byte("key")}
	// binary (non-UTF8) value
	value := &DeserializedData{data: []byte{0xff, 0xfe, 0xfd}}
	msg := &sarama.ConsumerMessage{}

	if filter.Matches(msg, key, value) {
		t.Fatal("expected binary non-UTF8 value to NOT match")
	}
}

func TestMessageFilter_BinaryKeyIgnored(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("test", "", nil)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	// binary (non-UTF8) key
	key := &DeserializedData{data: []byte{0xff, 0xfe, 0xfd}}
	value := &DeserializedData{data: []byte("value")}
	msg := &sarama.ConsumerMessage{}

	if filter.Matches(msg, key, value) {
		t.Fatal("expected binary non-UTF8 key to NOT match")
	}
}

func TestMessageFilter_HeaderMatch(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("", "", map[string]string{"trace-id": "abc-*"})
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	key := &DeserializedData{data: []byte("key")}
	value := &DeserializedData{data: []byte("value")}
	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("trace-id"), Value: []byte("abc-123")},
		},
	}

	if !filter.Matches(msg, key, value) {
		t.Fatal("expected header 'trace-id:abc-123' to match pattern 'abc-*'")
	}
}

func TestMessageFilter_HeaderNoMatch(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("", "", map[string]string{"trace-id": "abc-*"})
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	key := &DeserializedData{data: []byte("key")}
	value := &DeserializedData{data: []byte("value")}
	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("trace-id"), Value: []byte("xyz-456")},
		},
	}

	if filter.Matches(msg, key, value) {
		t.Fatal("expected header 'trace-id:xyz-456' to NOT match pattern 'abc-*'")
	}
}

func TestMessageFilter_HeaderMissing(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("", "", map[string]string{"trace-id": "abc-*"})
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	key := &DeserializedData{data: []byte("key")}
	value := &DeserializedData{data: []byte("value")}
	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("other-header"), Value: []byte("value")},
		},
	}

	if filter.Matches(msg, key, value) {
		t.Fatal("expected message without 'trace-id' header to NOT match")
	}
}

func TestMessageFilter_ANDLogic(t *testing.T) {
	t.Parallel()

	// All filters must match (AND logic)
	filter, err := NewMessageFilter("user-*", "*error*", map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	// All match
	key := &DeserializedData{data: []byte("user-123")}
	value := &DeserializedData{data: []byte("error occurred")}
	msg := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("env"), Value: []byte("prod")},
		},
	}

	if !filter.Matches(msg, key, value) {
		t.Fatal("expected all filters to match")
	}

	// Key doesn't match
	key2 := &DeserializedData{data: []byte("admin-456")}
	if filter.Matches(msg, key2, value) {
		t.Fatal("expected filter to fail when key doesn't match")
	}

	// Value doesn't match
	value2 := &DeserializedData{data: []byte("success")}
	if filter.Matches(msg, key, value2) {
		t.Fatal("expected filter to fail when value doesn't match")
	}

	// Header doesn't match
	msg2 := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{Key: []byte("env"), Value: []byte("dev")},
		},
	}
	if filter.Matches(msg2, key, value) {
		t.Fatal("expected filter to fail when header doesn't match")
	}
}

func TestMessageFilter_IsActive(t *testing.T) {
	t.Parallel()

	// No filters
	filter1, _ := NewMessageFilter("", "", nil)
	if filter1.IsActive() {
		t.Fatal("expected filter with no patterns to be inactive")
	}

	// Key filter only
	filter2, _ := NewMessageFilter("user-*", "", nil)
	if !filter2.IsActive() {
		t.Fatal("expected filter with key pattern to be active")
	}

	// Value filter only
	filter3, _ := NewMessageFilter("", "*error*", nil)
	if !filter3.IsActive() {
		t.Fatal("expected filter with value pattern to be active")
	}

	// Header filter only
	filter4, _ := NewMessageFilter("", "", map[string]string{"env": "prod"})
	if !filter4.IsActive() {
		t.Fatal("expected filter with header pattern to be active")
	}
}

func TestMessageFilter_InvalidGlobPattern(t *testing.T) {
	t.Parallel()

	// Invalid glob pattern (unclosed bracket)
	_, err := NewMessageFilter("[invalid", "", nil)
	if err == nil {
		t.Fatal("expected error for invalid glob pattern")
	}
}

func TestMessageFilter_NullKeyValue(t *testing.T) {
	t.Parallel()

	filter, err := NewMessageFilter("user-*", "", nil)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	// Nil key
	msg := &sarama.ConsumerMessage{}
	value := &DeserializedData{data: []byte("value")}

	if filter.Matches(msg, nil, value) {
		t.Fatal("expected filter to fail with nil key")
	}

	// Nil key data
	key := &DeserializedData{data: nil}
	if filter.Matches(msg, key, value) {
		t.Fatal("expected filter to fail with nil key data")
	}
}
