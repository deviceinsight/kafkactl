package get_test

import (
	"testing"

	"github.com/deviceinsight/kafkactl/internal/testutil"
)

func TestGetBrokersIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("get", "brokers"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := kafkaCtl.GetStdOutLines()

	testutil.AssertContains(t, "ID|ADDRESS", outputLines)
	testutil.AssertContains(t, "101|localhost:19093", outputLines)
	testutil.AssertContains(t, "102|localhost:29093", outputLines)
	testutil.AssertContains(t, "103|localhost:39093", outputLines)
}

func TestGetBrokersCompactIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("get", "brokers", "-o", "compact"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := kafkaCtl.GetStdOutLines()

	testutil.AssertContains(t, "localhost:19093", outputLines)
	testutil.AssertContains(t, "localhost:29093", outputLines)
	testutil.AssertContains(t, "localhost:39093", outputLines)
}

func TestGetBrokersSaslCompactIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("get", "brokers", "-o", "compact"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := kafkaCtl.GetStdOutLines()

	testutil.AssertContains(t, "localhost:19092", outputLines)
	testutil.AssertContains(t, "localhost:29092", outputLines)
	testutil.AssertContains(t, "localhost:39092", outputLines)
}
