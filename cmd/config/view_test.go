package config_test

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/test_util"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestViewConfigWithEnvVariablesInGeneratedConfigSet(t *testing.T) {

	test_util.StartUnitTest(t)

	currentDir, err := os.Getwd()

	if err != nil {
		t.Fatalf("unable to read current working dir: %v", err)
	}

	newConfigFile := path.Join(currentDir, "non-existing-config.yml")
	defer func() {
		if err = os.Remove(newConfigFile); err != nil {
			output.TestLogf("unable to delete file %s: %v", newConfigFile, err)
		}
	}()

	if err := os.Setenv("KAFKA_CTL_CONFIG", newConfigFile); err != nil {
		t.Fatalf("unable to set env variable: %v", err)
	}

	if err := os.Setenv("BROKERS", "env-broker:9092"); err != nil {
		t.Fatalf("unable to set env variable: %v", err)
	}

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	defaultConfigContent := `
contexts:
  default:
    brokers: env-broker:9092
current-context: default`

	if _, err := kafkaCtl.Execute("config", "view"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	configContent, err := ioutil.ReadFile(newConfigFile)
	if err != nil {
		t.Fatalf("error reading generated config %s %v", newConfigFile, err)
	}

	test_util.AssertEquals(t, defaultConfigContent, string(configContent))
	test_util.AssertEquals(t, defaultConfigContent, kafkaCtl.GetStdOut())
}
