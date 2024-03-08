package deletion_test

import (
	"fmt"
	"testing"

	"github.com/deviceinsight/kafkactl/internal/testutil"
)

func TestDeleteTopicReadAclIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.CreateTopic(t, "acl-topic")

	// add read acl
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// add write acl for other user
	// if no other acl is there everything is allowed after deletion
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "write", "--allow", "--principal", "User:admin"); err != nil {
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

	// switch to to sasl 'admin'
	testutil.SwitchContext("sasl-admin")

	// delete the acl
	if _, err := kafkaCtl.Execute("delete", "acl", "--topics", "--operation", "read", "--pattern", "any"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// switch to to sasl 'user'
	testutil.SwitchContext("sasl-user")

	// should not be able to consume
	_, err := kafkaCtl.Execute("consume", topicName, "--from-beginning", "--exit", "--print-keys")

	pre28ErrorMessage := fmt.Sprintf("topic '%s' does not exist", topicName)
	errorMessage := "The client is not authorized to access this topic"

	testutil.AssertErrorContainsOneOf(t, []string{pre28ErrorMessage, errorMessage}, err)
}
