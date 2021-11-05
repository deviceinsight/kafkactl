package get_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/testutil"
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
