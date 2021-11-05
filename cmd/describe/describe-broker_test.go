package describe_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/internal/broker"

	"github.com/deviceinsight/kafkactl/testutil"
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
