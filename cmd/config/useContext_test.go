package config_test

import (
	"github.com/deviceinsight/kafkactl/test_util"
	"strings"
	"testing"
)

func TestUseContextAutoCompletionIntegration(t *testing.T) {

	test_util.StartUnitTest(t)

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "config", "use-context", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	expectedContexts := []string{"default", "no-avro", "sasl-admin", "sasl-user"}

	if len(outputLines) != len(expectedContexts)+1 {
		t.Fatalf("unexpected output. expected %d lines got %d: %s", len(expectedContexts)+1, len(outputLines), kafkaCtl.GetStdOut())
	}

	for _, context := range expectedContexts {
		test_util.AssertContains(t, context, outputLines)
	}
}
