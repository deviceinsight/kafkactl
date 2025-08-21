package alter_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestAlterBrokerAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "alter", "broker", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	testutil.AssertArraysEquals(t, []string{"101", "102", "103", ":4"}, outputLines)
}

func TestAlterBrokerConfigIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "broker", "101", "--config", "background.threads=10"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// alter another config to ensure config merge is working
	if _, err := kafkaCtl.Execute("alter", "broker", "101", "--config", "log.cleaner.threads=2"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	if _, err := kafkaCtl.Execute("alter", "broker", "101", "--config", "auto.create.topics.enable=false", "--validate-only"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	config1Regex, _ := regexp.Compile(`background.threads\s+10\s`)
	if config1Regex.FindString(output) == "" {
		t.Fatalf("expected output to contain 'background.threads 10', got: %s", output)
	}

	config2Regex, _ := regexp.Compile(`log.cleaner.threads\s+2\s`)
	if config2Regex.FindString(output) == "" {
		t.Fatalf("expected output to contain 'log.cleaner.threads 2', got: %s", output)
	}
}

func TestAlterBrokerValidateOnlyIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "broker", "101", "--config", "background.threads=20", "--validate-only"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	config1Regex, _ := regexp.Compile(`background.threads\s+20\s`)
	if config1Regex.FindString(output) == "" {
		t.Fatalf("expected output to contain 'background.threads 20', got: %s", output)
	}
}

func TestAlterBrokerNonDynamicConfigIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	_, err := kafkaCtl.Execute("alter", "broker", "101", "--config", "broker.id=999")
	if err == nil {
		t.Fatalf("expected error when trying to alter non-dynamic config broker.id, but command succeeded")
	}

	if !strings.Contains(err.Error(), "Cannot update these configs dynamically") {
		t.Fatalf("expected error about non-dynamic configs, got: %v", err)
	}

	if !strings.Contains(err.Error(), "broker.id") {
		t.Fatalf("expected broker.id to be mentioned in non-dynamic config error, got: %v", err)
	}
}

func TestAlterBrokerClusterWideConfigIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "broker", "", "--config", "background.threads=8"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	if !strings.Contains(output, "config has been altered") {
		t.Fatalf("expected 'config has been altered' message, got: %s", output)
	}

	// Note: We don't verify the actual config values here because cluster-wide configs
	// behave differently than individual broker configs and may not immediately show
	// on all brokers depending on their individual override status
}

func TestAlterBrokerClusterWideValidateOnlyIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "broker", "", "--config", "background.threads=12", "--validate-only"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	config1Regex, _ := regexp.Compile(`background.threads\s+12\s`)
	if config1Regex.FindString(output) == "" {
		t.Fatalf("expected output to contain 'background.threads 12', got: %s", output)
	}
}

func TestAlterBrokerConfigK8sIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "k8s-mock")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	type testCases struct {
		description      string
		args             []string
		wantInKubectlCmd []string
	}

	for _, test := range []testCases{
		{
			description:      "single_config_defined_with_space",
			args:             []string{"alter", "broker", "1", "--config", "background.threads=10"},
			wantInKubectlCmd: []string{"--config=background.threads=10"},
		},
		{
			description:      "single_config_defined_with_equal",
			args:             []string{"alter", "broker", "1", "--config=background.threads=10"},
			wantInKubectlCmd: []string{"--config=background.threads=10"},
		},
		{
			description: "multiple_configs",
			args: []string{"alter", "broker", "1", "--config", "background.threads=10",
				"--config", "log.cleaner.threads=1"},
			wantInKubectlCmd: []string{"--config=background.threads=10", "--config=log.cleaner.threads=1"},
		},
		{
			description:      "validate_only_flag",
			args:             []string{"alter", "broker", "1", "--config", "background.threads=10", "--validate-only"},
			wantInKubectlCmd: []string{"--config=background.threads=10", "--validate-only"},
		},
	} {
		t.Run(test.description, func(t *testing.T) {

			if _, err := kafkaCtl.Execute(test.args...); err != nil {
				t.Fatalf("failed to execute command: %v", err)
			}

			output := kafkaCtl.GetStdOut()

			for _, wanted := range test.wantInKubectlCmd {
				testutil.AssertContainSubstring(t, wanted, output)
			}
		})
	}
}
