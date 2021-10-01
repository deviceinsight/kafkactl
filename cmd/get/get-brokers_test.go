package get_test

import (
	"github.com/deviceinsight/kafkactl/test_util"
	"strings"
	"testing"
)

func TestGetBrokersIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("get", "brokers"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, "101     localhost:19093", outputLines)
	test_util.AssertContains(t, "102     localhost:29093", outputLines)
	test_util.AssertContains(t, "103     localhost:39093", outputLines)
}

func TestGetBrokersCompactIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("get", "brokers", "-o", "compact"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, "localhost:19093", outputLines)
	test_util.AssertContains(t, "localhost:29093", outputLines)
	test_util.AssertContains(t, "localhost:39093", outputLines)
}

func TestGetBrokersSaslCompactIntegration(t *testing.T) {

	test_util.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("get", "brokers", "-o", "compact"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, "localhost:19092", outputLines)
	test_util.AssertContains(t, "localhost:29092", outputLines)
	test_util.AssertContains(t, "localhost:39092", outputLines)
}
