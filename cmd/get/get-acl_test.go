package get_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/acl"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestGetTopicReadAclIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.CreateTopic(t, "acl-topic")

	// add read acl
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// add write acl for other user
	// if no other acl is there everything is allowed after deletion
	if _, err := kafkaCtl.Execute("get", "acl", "--topics", "--operation", "read"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	re := regexp.MustCompile(`\s+`)
	outputLines := make([]string, 0)

	for _, line := range strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n") {
		outputLines = append(outputLines, re.ReplaceAllString(line, " "))
	}

	testutil.AssertContains(t, fmt.Sprintf("Topic %s Literal User:user * Read Allow", topicName), outputLines)
}

func TestGetAclByPrincipalIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.CreateTopic(t, "acl-get-topic-principal")

	// add read acl
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// add write acl for other user
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:admin"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// list acls
	if _, err := kafkaCtl.Execute("get", "acl", "--resource-name", topicName, "--principal", "User:user", "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	acls, err := acl.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}
	testutil.AssertIntEquals(t, 1, len(acls))
	testutil.AssertIntEquals(t, 1, len(acls[0].Acls))
	testutil.AssertEquals(t, "User:user", acls[0].Acls[0].Principal)
}

func TestGetAclByHostIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	topicName := testutil.CreateTopic(t, "acl-get-topic-host")

	// add acl for host-a
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user", "--host", "host-a"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// add acl for host-b
	if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user", "--host", "host-b"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	// list acls
	if _, err := kafkaCtl.Execute("get", "acl", "--resource-name", topicName, "--host", "host-a", "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	acls, err := acl.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}
	testutil.AssertIntEquals(t, 1, len(acls))
	testutil.AssertIntEquals(t, 1, len(acls[0].Acls))
	testutil.AssertEquals(t, "host-a", acls[0].Acls[0].Host)
}
