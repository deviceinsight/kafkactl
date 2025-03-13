package schemaregistry

import (
	"strings"
	"testing"
)

func TestFormatBaseURL(t *testing.T) {
	testCases := []struct {
		userConfig string
		expected   string
	}{
		{"localhost", "http://localhost:80"},
		{"localhost:443", "https://localhost:443"},
		{"https://localhost", "https://localhost:443"},
		{"localhost:8080/", "http://localhost:8080"},
		{"https://schema.registry.com/apis/ccompat/v7/", "https://schema.registry.com:443/apis/ccompat/v7"},
	}
	for _, tc := range testCases {
		t.Run(tc.userConfig, func(t *testing.T) {
			actual := FormatBaseURL(tc.userConfig)
			if strings.TrimSpace(actual) != strings.TrimSpace(tc.expected) {
				t.Fatalf("unexpected output:\nexpected:\n--\n%s\n--\nactual:\n--\n%s\n--", tc.expected, strings.TrimSpace(actual))
			}
		})
	}
}
