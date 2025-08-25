package config_test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestViewConfigWithEnvVariablesInGeneratedConfigSet(t *testing.T) {

	testutil.StartUnitTest(t)

	currentDir, err := os.Getwd()

	if err != nil {
		t.Fatalf("unable to read current working dir: %v", err)
	}

	newConfigFile := path.Join(currentDir, "non-existing-config.yml")
	newContextFile := path.Join(currentDir, "non-existing-context.yml")
	defer func() {
		if err = os.Remove(newConfigFile); err != nil {
			output.TestLogf("unable to delete file %s: %v", newConfigFile, err)
		}
		if err = os.Remove(newContextFile); err != nil {
			output.TestLogf("unable to delete file %s: %v", newContextFile, err)
		}
	}()

	if err := os.Setenv("KAFKA_CTL_CONFIG", newConfigFile); err != nil {
		t.Fatalf("unable to set env variable: %v", err)
	}

	if err := os.Setenv("KAFKA_CTL_WRITABLE_CONFIG", newContextFile); err != nil {
		t.Fatalf("unable to set env variable: %v", err)
	}

	if err := os.Setenv("BROKERS", "env-broker:9092"); err != nil {
		t.Fatalf("unable to set env variable: %v", err)
	}

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	defaultConfigContent := `
contexts:
    default:
        brokers: env-broker:9092`

	defaultContextContent := `
current-context: default`

	if _, err := kafkaCtl.Execute("config", "view"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	configContent, err := os.ReadFile(newConfigFile)
	if err != nil {
		t.Fatalf("error reading generated config %s %v", newConfigFile, err)
	}

	contextContent, err := os.ReadFile(newContextFile)
	if err != nil {
		t.Fatalf("error reading generated config %s %v", newContextFile, err)
	}

	testutil.AssertEquals(t, defaultConfigContent, string(configContent))
	testutil.AssertEquals(t, defaultContextContent, string(contextContent))
	testutil.AssertEquals(t, defaultConfigContent, kafkaCtl.GetStdOut())
}

func TestViewConfigWorksIfConfigDirIsReadOnly(t *testing.T) {

	tests := []struct {
		name             string
		configFileExists bool
		wantError        string
	}{
		{
			name:      "config_and_context_file_cannot_be_created",
			wantError: "no config file loaded",
		},
		{
			name:             "config_file_exists_but_context_file_cannot_be_created",
			configFileExists: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testutil.StartUnitTest(t)

			tempDir := getTempDir()
			if err := os.MkdirAll(tempDir, 0755); err != nil {
				t.Fatalf("unable to create temp dir: %v", err)
			}
			defer func(path string) {
				err := os.RemoveAll(path)
				if err != nil {
					output.TestLogf("unable to delete temp dir %s: %v", path, err)
				}
			}(tempDir)

			configFile := path.Join(tempDir, "config.yml")
			contextFile := path.Join(tempDir, "context.yml")

			configContent := `
contexts:
    default:
        brokers: env-broker:9092`

			if tt.configFileExists {
				if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
					t.Fatalf("unable to create config file: %v", err)
				}
			}

			if err := os.Chmod(tempDir, 0555); err != nil {
				t.Fatalf("unable to make config file readonly: %v", err)
			}

			if err := os.Setenv("KAFKA_CTL_CONFIG", configFile); err != nil {
				t.Fatalf("unable to set env variable: %v", err)
			}

			if err := os.Setenv("KAFKA_CTL_WRITABLE_CONFIG", contextFile); err != nil {
				t.Fatalf("unable to set env variable: %v", err)
			}

			if err := os.Setenv("BROKERS", "env-broker:9092"); err != nil {
				t.Fatalf("unable to set env variable: %v", err)
			}

			kafkaCtl := testutil.CreateKafkaCtlCommand()

			_, err := kafkaCtl.Execute("config", "view")

			if tt.wantError == "" && err != nil {
				t.Fatalf("failed to execute command: %v", err)
			} else if tt.wantError != "" && err == nil {
				t.Fatalf("expecting error but got none: %s", tt.wantError)
				return
			} else if tt.wantError != "" {
				testutil.AssertContainSubstring(t, tt.wantError, err.Error())
				return
			}

			actualConfigContent, err := os.ReadFile(configFile)
			if err != nil {
				t.Fatalf("error reading generated config %s %v", configFile, err)
			}

			if _, err := os.Stat(contextFile); err == nil {
				t.Fatalf("context file should not exist %s: %v", contextFile, err)
			}

			testutil.AssertEquals(t, configContent, string(actualConfigContent))
			testutil.AssertEquals(t, configContent, kafkaCtl.GetStdOut())
			testutil.AssertContainSubstring(t, "cannot write config file", kafkaCtl.GetStdErr())
		})
	}
}

func getTempDir() string {
	timestamp := time.Now().UnixNano()
	dirName := fmt.Sprintf("test-dir-%d", timestamp)
	return filepath.Join(os.TempDir(), dirName)
}
