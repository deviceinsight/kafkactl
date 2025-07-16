package deletion_test

import (
	"fmt"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/acl"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
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

func TestDeleteAclByPrincipalIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.CreateTopic(t, "acl-topic-principal")

	// add read acl
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// add write acl for other user
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:admin"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// list acls
	if _, err := kafkaCtl.Execute("get", "acl", "--resource-name", topicName, "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	acls, err := acl.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}
	testutil.AssertIntEquals(t, 1, len(acls))
	testutil.AssertIntEquals(t, 2, len(acls[0].Acls))
	principals := []string{acls[0].Acls[0].Principal, acls[0].Acls[1].Principal}
	testutil.AssertContains(t, "User:user", principals)
	testutil.AssertContains(t, "User:admin", principals)

	// delete the acl
	if _, err := kafkaCtl.Execute("delete", "acl", "--topics", "--operation", "read", "--principal", "User:user", "--pattern", "literal"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// list acls
	if _, err := kafkaCtl.Execute("get", "acl", "--resource-name", topicName, "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	acls, err = acl.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}
	testutil.AssertIntEquals(t, 1, len(acls))
	testutil.AssertIntEquals(t, 1, len(acls[0].Acls))
	testutil.AssertEquals(t, "User:admin", acls[0].Acls[0].Principal)
}

func TestDeleteAclByHostIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.CreateTopic(t, "acl-topic-host")

	// add acl for host-a
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user", "--host", "host-a"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// add acl for host-b
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user", "--host", "host-b"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// list acls
	if _, err := kafkaCtl.Execute("get", "acl", "--resource-name", topicName, "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	acls, err := acl.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}
	testutil.AssertIntEquals(t, 1, len(acls))
	testutil.AssertIntEquals(t, 2, len(acls[0].Acls))

	hosts := []string{acls[0].Acls[0].Host, acls[0].Acls[1].Host}
	testutil.AssertContains(t, "host-a", hosts)
	testutil.AssertContains(t, "host-b", hosts)

	// delete the acl
	if _, err := kafkaCtl.Execute("delete", "acl", "--topics", "--operation", "read", "--host", "host-a", "--pattern", "literal"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// list acls
	if _, err := kafkaCtl.Execute("get", "acl", "--resource-name", topicName, "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	acls, err = acl.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}
	testutil.AssertIntEquals(t, 1, len(acls))
	testutil.AssertIntEquals(t, 1, len(acls[0].Acls))
	testutil.AssertEquals(t, "host-b", acls[0].Acls[0].Host)
}
