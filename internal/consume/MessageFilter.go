package consume

import (
	"fmt"
	"unicode/utf8"

	"github.com/IBM/sarama"
	"github.com/gobwas/glob"
)

type MessageFilter struct {
	keyGlob    glob.Glob
	valueGlob  glob.Glob
	headerGlob map[string]glob.Glob
}

func NewMessageFilter(filterKey, filterValue string, filterHeader map[string]string) (*MessageFilter, error) {
	filter := &MessageFilter{}
	var err error

	if filterKey != "" {
		filter.keyGlob, err = glob.Compile(filterKey)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern for --filter-key: %w", err)
		}
	}

	if filterValue != "" {
		filter.valueGlob, err = glob.Compile(filterValue)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern for --filter-value: %w", err)
		}
	}

	if len(filterHeader) > 0 {
		filter.headerGlob = make(map[string]glob.Glob)
		for key, pattern := range filterHeader {
			filter.headerGlob[key], err = glob.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid glob pattern for --filter-header %s: %w", key, err)
			}
		}
	}

	return filter, nil
}

func (f *MessageFilter) Matches(consumerMsg *sarama.ConsumerMessage, key, value *DeserializedData) bool {
	// Filter by key
	if f.keyGlob != nil {
		if key == nil || key.data == nil || !utf8.Valid(key.data) {
			return false
		}
		if !f.keyGlob.Match(string(key.data)) {
			return false
		}
	}

	// Filter by value
	if f.valueGlob != nil {
		if value == nil || value.data == nil || !utf8.Valid(value.data) {
			return false
		}
		if !f.valueGlob.Match(string(value.data)) {
			return false
		}
	}

	// Filter by headers
	if len(f.headerGlob) > 0 {
		headers := encodeRecordHeaders(consumerMsg.Headers)
		for headerKey, globPattern := range f.headerGlob {
			headerValue, exists := headers[headerKey]
			if !exists {
				return false
			}
			if !globPattern.Match(headerValue) {
				return false
			}
		}
	}

	return true
}

func (f *MessageFilter) IsActive() bool {
	return f.keyGlob != nil || f.valueGlob != nil || len(f.headerGlob) > 0
}
