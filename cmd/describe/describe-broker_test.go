package describe_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/broker"
	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestDescribeBrokerIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("describe", "broker", "101", "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	broker, err := broker.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("could not convert output to broker yaml: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprint(broker.ID), "101")

	expectedConfigs := map[string]string{
		"auto.create.topics.enable": "false",
	}

	for _, config := range broker.Configs {

		if expectedValue, ok := expectedConfigs[config.Name]; ok {
			testutil.AssertEquals(t, config.Value, expectedValue)
			delete(expectedConfigs, config.Name)
		}
	}

	if len(expectedConfigs) > 0 {
		t.Fatalf("expected configs missing: %v", expectedConfigs)
	}
}

func TestDescribeBrokerConfigsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("describe", "broker", "101", "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	describedBroker, err := broker.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}

	configKeys := getConfigKeys(describedBroker.Configs)
	if len(configKeys) < 10 {
		t.Fatalf("expected to find >=10 config keys, found %v", len(configKeys))
	}

	if _, err := kafkaCtl.Execute("describe", "broker", "101", "--all-configs", "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	describedBroker, err = broker.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}

	allConfigKeys := getConfigKeys(describedBroker.Configs)

	// Verify allConfigKeys contains more entries than configKeys
	if len(allConfigKeys) <= len(configKeys) {
		t.Fatalf("expected allConfigKeys (%d) to contain more entries than configKeys (%d)",
			len(allConfigKeys), len(configKeys))
	}

	// Verify all strings in configKeys are also in allConfigKeys
	allConfigKeysMap := make(map[string]bool)
	for _, key := range allConfigKeys {
		allConfigKeysMap[key] = true
	}

	for _, key := range configKeys {
		if !allConfigKeysMap[key] {
			t.Fatalf("config key %q found in configKeys but not in allConfigKeys", key)
		}
	}
}

func TestDescribeBrokerAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "describe", "broker", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, "101", outputLines)
	testutil.AssertContains(t, "102", outputLines)
	testutil.AssertContains(t, "103", outputLines)
}
