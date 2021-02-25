package create_test

import (
	"fmt"
	"github.com/deviceinsight/kafkactl/test_util"
	"testing"
)

func TestCreateTopicReadAclIntegration(t *testing.T) {

	test_util.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	topicName := test_util.CreateTopic(t, "acl-topic")

	// add read acl
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// produce to topic
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	// switch to to sasl 'user'
	test_util.SwitchContext("sasl-user")

	// user should be able to consume
	if _, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "test-key#test-value", kafkaCtl.GetStdOut())

	// but cannot produce
	_, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value")
	test_util.AssertErrorContains(t, "The client is not authorized to access this topic", err)
}

func TestCreateTopicWriteAclIntegration(t *testing.T) {

	test_util.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	topicName := test_util.CreateTopic(t, "acl-topic")

	// deny read acl
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--deny", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// produce to topic
	if _, err := kafkaCtl.Execute("produce", topicName, "--key", "test-key", "--value", "test-value"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, "message produced (partition=0\toffset=0)", kafkaCtl.GetStdOut())

	// switch to to sasl 'user'
	test_util.SwitchContext("sasl-user")

	// should not be able to consume
	_, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys")
	test_util.AssertErrorContains(t, fmt.Sprintf("topic '%s' does not exist", topicName), err)
}

func TestCreateTopicAlterConfigsDenyAclIntegration(t *testing.T) {

	test_util.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	topicName := test_util.CreateTopic(t, "acl-topic")

	// deny alterConfigs
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "alterConfigs", "--deny", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// add read acl
	kafkaCtl = test_util.CreateKafkaCtlCommand()
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	if _, err := kafkaCtl.Execute("alter", "topic", topicName, "--config", "retention.ms=3600000"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// switch to to sasl 'user'
	test_util.SwitchContext("sasl-user")

	_, err := kafkaCtl.Execute("alter", "topic", topicName, "--config", "retention.ms=600000")
	test_util.AssertErrorContains(t, "The client is not authorized to access this topic", err)
}
