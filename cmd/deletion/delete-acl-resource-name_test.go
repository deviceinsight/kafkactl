package deletion_test

import (
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal/acl"
	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestDeleteAclByResourceNameIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	selectedTopic := testutil.CreateTopic(t, "acl-topic-selected")
	remainingTopic := testutil.CreateTopic(t, "acl-topic-remaining")

	for _, topicName := range []string{selectedTopic, remainingTopic} {
		if _, err := kafkaCtl.Execute("create", "acl", "--topic", topicName, "--operation", "read", "--allow", "--principal", "User:user"); err != nil {
			t.Fatalf("failed to execute command: %v", err)
		}
	}

	if _, err := kafkaCtl.Execute("delete", "acl", "--topics", "--resource-name", selectedTopic, "--operation", "read", "--pattern", "literal"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	if _, err := kafkaCtl.Execute("get", "acl", "--resource-name", selectedTopic, "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	entries, err := acl.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}
	testutil.AssertIntEquals(t, 0, len(entries))

	if _, err := kafkaCtl.Execute("get", "acl", "--resource-name", remainingTopic, "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	entries, err = acl.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}
	testutil.AssertIntEquals(t, 1, len(entries))
	testutil.AssertEquals(t, remainingTopic, entries[0].ResourceName)
	testutil.AssertIntEquals(t, 1, len(entries[0].Acls))
}
