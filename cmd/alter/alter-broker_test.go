package alter_test

import (
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

	// Check that we get broker IDs in the completion (usually 101, 102, 103 in test environment)
	// We can't predict exact broker IDs, but we should get some numeric IDs
	if len(outputLines) == 0 {
		t.Fatalf("expected broker IDs in autocompletion, got empty result")
	}

	// Verify that at least one line looks like a broker ID (numeric, may have shell completion prefix)
	foundNumericID := false
	for _, line := range outputLines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Remove shell completion directive prefix if present (e.g., ":1" -> "1")
			line = strings.TrimPrefix(line, ":")
			if isNumeric(line) {
				foundNumericID = true
				break
			}
		}
	}

	if !foundNumericID {
		t.Fatalf("expected at least one numeric broker ID in autocompletion, got: %v", outputLines)
	}
}

func TestAlterBrokerConfigIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	// Get first broker ID
	brokerID := getFirstBrokerID(t, kafkaCtl)

	// Test altering a broker config - this tests the full command path
	// Even if config isn't dynamically configurable, we verify proper error handling
	_, err := kafkaCtl.Execute("alter", "broker", brokerID, "--config", "background.threads=10")

	if err != nil {
		// If it fails due to non-dynamic config, that's expected - verify proper error
		if strings.Contains(err.Error(), "Cannot update these configs dynamically") {
			// This is the expected behavior for non-dynamic configs
			t.Logf("Config not dynamically configurable (expected): %v", err)
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	} else {
		// If it succeeds, verify the success message
		output := kafkaCtl.GetStdOut()
		if !strings.Contains(output, "config has been altered") {
			t.Fatalf("expected 'config has been altered' message, got: %s", output)
		}
	}
}

func TestAlterBrokerValidateOnlyIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	// Get first broker ID
	brokerID := getFirstBrokerID(t, kafkaCtl)

	// Test validate-only flag - even if config isn't dynamically changeable,
	// validate-only should not show "config has been altered" message
	_, err := kafkaCtl.Execute("alter", "broker", brokerID, "--config", "background.threads=10", "--validate-only")

	output := kafkaCtl.GetStdOut()

	// Validate-only should never show "config has been altered" message, even on error
	if strings.Contains(output, "config has been altered") {
		t.Fatalf("validate-only should not show 'config has been altered' message, got: %s", output)
	}

	if err != nil {
		t.Fatalf("unexpected error in validate-only mode: %v", err)
	}

	// With validate-only, the config output goes directly to stdout via fmt.Println
	// The test framework captures this output, so we don't check GetStdOut() here
	// The fact that there was no error means validate-only worked correctly
}

func TestAlterBrokerNonDynamicConfigIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	// Get first broker ID
	brokerID := getFirstBrokerID(t, kafkaCtl)

	// Test attempting to alter a non-dynamic config - should fail with proper error
	_, err := kafkaCtl.Execute("alter", "broker", brokerID, "--config", "broker.id=999")

	if err == nil {
		t.Fatalf("expected error when trying to alter non-dynamic config broker.id, but command succeeded")
	}

	// Verify we get the expected error about non-dynamic configs
	if !strings.Contains(err.Error(), "Cannot update these configs dynamically") {
		t.Fatalf("expected error about non-dynamic configs, got: %v", err)
	}

	// Verify broker.id is mentioned in the error (it should be in the list of non-dynamic configs)
	if !strings.Contains(err.Error(), "broker.id") {
		t.Fatalf("expected broker.id to be mentioned in non-dynamic config error, got: %v", err)
	}
}

func TestAlterBrokerClusterWideConfigIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	// Test cluster-wide config alteration with empty broker ID
	_, err := kafkaCtl.Execute("alter", "broker", "", "--config", "background.threads=8")

	if err != nil {
		t.Fatalf("failed to alter cluster-wide broker config: %v", err)
	}

	// Verify the success message
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

	// Test cluster-wide validate-only with empty broker ID (use a valid value within range)
	_, err := kafkaCtl.Execute("alter", "broker", "", "--config", "background.threads=12", "--validate-only")

	if err != nil {
		t.Fatalf("unexpected error in cluster-wide validate-only mode: %v", err)
	}

	// Validate-only should not show "config has been altered" message
	output := kafkaCtl.GetStdOut()
	if strings.Contains(output, "config has been altered") {
		t.Fatalf("validate-only should not show 'config has been altered' message, got: %s", output)
	}

	// With cluster-wide validate-only, we should see current cluster-wide config values
	// The exact output format may vary, so we just verify no error occurred
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

// Helper function to check if a string is numeric
func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Helper function to get the first available broker ID
func getFirstBrokerID(t *testing.T, kafkaCtl testutil.KafkaCtlTestCommand) string {
	if _, err := kafkaCtl.Execute("get", "brokers"); err != nil {
		t.Fatalf("failed to get brokers: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	lines := strings.Split(output, "\n")

	// Skip header line, look for first data line
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header or empty lines
		}
		fields := strings.Fields(line)
		if len(fields) > 0 && isNumeric(fields[0]) {
			return fields[0]
		}
	}

	t.Fatalf("could not find any broker ID in output: %s", output)
	return ""
}
