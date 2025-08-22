package config_test

import (
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestUseContextAutoCompletionIntegration(t *testing.T) {

	testutil.StartUnitTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "config", "use-context", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	expectedContexts := []string{"default", "k8s-mock", "no-avro", "sasl-admin", "sasl-user", "scram-admin"}

	if len(outputLines) != len(expectedContexts)+1 {
		t.Fatalf("unexpected output. expected %d lines got %d: %s", len(expectedContexts)+1, len(outputLines), kafkaCtl.GetStdOut())
	}

	for _, context := range expectedContexts {
		testutil.AssertContains(t, context, outputLines)
	}
}
