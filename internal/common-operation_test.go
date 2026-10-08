package internal

import (
	"encoding/pem"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/credential"
	"github.com/deviceinsight/kafkactl/v5/internal/global"
	"github.com/spf13/viper"
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

func TestResolvePassphraseIgnoresLocalKeyWithKubernetes(t *testing.T) {
	encryptedKey := filepath.Join(t.TempDir(), "tls.key")
	if err := os.WriteFile(encryptedKey, pem.EncodeToMemory(&pem.Block{Type: "ENCRYPTED PRIVATE KEY"}), 0o600); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name               string
		kubernetes         bool
		certKey            string
		passphrase         string
		expectedPassphrase string
		expectedError      string
	}{
		{name: "local missing key", certKey: "missing/tls.key", expectedError: "unable to read missing/tls.key"},
		{name: "local encrypted key", certKey: encryptedKey, expectedError: "no terminal available for prompting"},
		{name: "kubernetes missing key", kubernetes: true, certKey: "missing/tls.key"},
		{name: "kubernetes encrypted local key", kubernetes: true, certKey: encryptedKey},
		{name: "kubernetes with configured passphrase", kubernetes: true, certKey: encryptedKey, passphrase: "secret", expectedPassphrase: "secret"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(viper.Reset)
			global.NewConfig().Flags().Context = "test"
			viper.Set("contexts.test.kubernetes.enabled", tc.kubernetes)
			if tc.passphrase != "" {
				viper.Set("contexts.test.tls.certKeyPassphrase", tc.passphrase)
			}

			passphrase, err := resolvePassphrase(credential.NewPromptCredentialResolver(), "test", tc.certKey, "tls.certKeyPassphrase", "label")

			if tc.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Fatalf("expected error containing %q, got: %v", tc.expectedError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if passphrase != tc.expectedPassphrase {
				t.Fatalf("expected passphrase %q, got %q", tc.expectedPassphrase, passphrase)
			}
		})
	}
}
