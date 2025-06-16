package create_test

import (
	"fmt"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestCreateTopicReadAclIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.CreateTopic(t, "acl-topic")

	// add read acl
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// produce to topic
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	// switch to to sasl 'user'
	testutil.SwitchContext("sasl-user")

	// user should be able to consume
	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "test-key#test-value", kafkaCtl.GetStdOut())

	// but cannot produce
	_, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value")
	testutil.AssertErrorContains(t, "The client is not authorized to access this topic", err)
}

func TestCreateTopicWriteAclIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.CreateTopic(t, "acl-topic")

	// deny read acl
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--deny", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// produce to topic
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	// switch to to sasl 'user'
	testutil.SwitchContext("sasl-user")

	// should not be able to consume
	_, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys")

	pre28ErrorMessage := fmt.Sprintf("topic '%s' does not exist", topicName)
	errorMessage := "The client is not authorized to access this topic"

	testutil.AssertErrorContainsOneOf(t, []string{pre28ErrorMessage, errorMessage}, err)
}

func TestCreateTopicAlterConfigsDenyAclIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.CreateTopic(t, "acl-topic")

	// deny alterConfigs
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "alterConfigs", "--deny", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// add read acl
	kafkaCtl = testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	if _, err := kafkaCtl.Execute("alter", "topic", topicName, "--config", "retention.ms=3600000"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// switch to to sasl 'user'
	testutil.SwitchContext("sasl-user")

	_, err := kafkaCtl.Execute("alter", "topic", topicName, "--config", "retention.ms=600000")

	pre27ErrorMessage := "The client is not authorized to access this topic"
	errorMessage := "Topic authorization failed"

	testutil.AssertErrorContainsOneOf(t, []string{pre27ErrorMessage, errorMessage}, err)
}

func TestCreateClusterIdempotentWriteAllowAclIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	// add cluster acl
	kafkaCtl := testutil.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("create", "acl", "--cluster", "--operation", "IdempotentWrite", "--allow", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// remove acl again, it has side effects on other tests
	if _, err := kafkaCtl.Execute("delete", "acl", "--cluster", "--operation", "IdempotentWrite", "--pattern", "any"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
}
