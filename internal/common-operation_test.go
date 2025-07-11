package internal

import (
	"reflect"
	"testing"

	"github.com/IBM/sarama"
)

func TestListConfigsFromEntries(t *testing.T) {
	testCases := []struct {
		name            string
		entries         []sarama.ConfigEntry
		includeDefaults bool
		configs         []Config
	}{
		{
			name:    "not include defaults, empty entries",
			entries: []sarama.ConfigEntry{},
			configs: []Config{},
		},
		{
			name: "not include defaults",
			entries: []sarama.ConfigEntry{
				{
					Name:    "non_default",
					Value:   "ND",
					Default: false,
					Source:  sarama.SourceUnknown,
				},
				{
					Name:    "default",
					Value:   "D",
					Default: true,
					Source:  sarama.SourceDefault,
				},
			},
			configs: []Config{
				{Name: "non_default", Value: "ND"},
			},
		},
		{
			name:            "include defaults, empty entries",
			entries:         []sarama.ConfigEntry{},
			configs:         []Config{},
			includeDefaults: true,
		},
		{
			name: "include defaults",
			entries: []sarama.ConfigEntry{
				{
					Name:    "non_default",
					Value:   "ND",
					Default: false,
					Source:  sarama.SourceUnknown,
				},
				{
					Name:    "default",
					Value:   "D",
					Default: true,
					Source:  sarama.SourceDefault,
				},
			},
			configs: []Config{
				{Name: "non_default", Value: "ND"},
				{Name: "default", Value: "D"},
			},
			includeDefaults: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configs := listConfigsFromEntries(tc.entries, tc.includeDefaults)

			if len(configs) > 0 &&
				len(tc.configs) > 0 &&
				!reflect.DeepEqual(configs, tc.configs) {
				t.Fatalf("expect: %v, got %v", tc.configs, configs)
			}
		})
	}
}

func TestSanitizeUsername(t *testing.T) {
	testCases := []struct {
		descriptions string
		username     string
		want         string
	}{
		{
			descriptions: "windows user with domain",
			username:     "DOMAIN|MACHINE\\username",
			want:         "username",
		},
		{
			descriptions: "user with email",
			username:     "user@domain.com",
			want:         "user-domain-com",
		},
		{
			descriptions: "user with underscores",
			username:     "user_with__underscores",
			want:         "user-with-underscores",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.descriptions, func(t *testing.T) {
			if tc.want != sanitizeUsername(tc.username) {
				t.Fatalf("expected:\n--\n%s\n--\nactual:\n--\n%s\n--", tc.want, sanitizeUsername(tc.username))
			}
		})
	}
}
