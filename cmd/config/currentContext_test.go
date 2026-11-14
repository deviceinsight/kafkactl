package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestCurrentContextIsStoredNextToConfigFileFromFlag(t *testing.T) {

	testutil.StartUnitTest(t)

	homeDir := t.TempDir()
	configDir := t.TempDir()

	t.Setenv("HOME", homeDir)
	t.Setenv("KAFKA_CTL_CONFIG", "")
	t.Setenv("KAFKA_CTL_WRITABLE_CONFIG", "")

	// a current context left behind in the default location must not be used
	defaultContextDir := filepath.Join(homeDir, ".config", "kafkactl")
	if err := os.MkdirAll(defaultContextDir, 0700); err != nil {
		t.Fatalf("unable to create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(defaultContextDir, "current-context.yml"), []byte("current-context: default\n"), 0600); err != nil {
		t.Fatalf("unable to write file: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yml")
	if err := os.WriteFile(configFile, []byte("contexts:\n  local:\n    brokers: [localhost:9092]\n"), 0600); err != nil {
		t.Fatalf("unable to write file: %v", err)
	}

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("config", "current-context", "-C", configFile); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "local", strings.TrimSpace(kafkaCtl.GetStdOut()))

	contextContent, err := os.ReadFile(filepath.Join(configDir, "current-context.yml"))
	if err != nil {
		t.Fatalf("current context was not stored next to config file: %v", err)
	}

	testutil.AssertEquals(t, "current-context: local\n", string(contextContent))
}
