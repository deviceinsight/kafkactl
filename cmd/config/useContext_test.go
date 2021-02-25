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

	if len(outputLines) != 4 {
		t.Fatalf("unexpected output. expected two lines got %d: %s", len(outputLines), kafkaCtl.GetStdOut())
	}

	test_util.AssertEquals(t, "default", outputLines[0])
	test_util.AssertEquals(t, "sasl-admin", outputLines[1])
	test_util.AssertEquals(t, "sasl-user", outputLines[2])
}
