package get_test

import (
	"fmt"
	"github.com/deviceinsight/kafkactl/test_util"
	"regexp"
	"strings"
	"testing"
)

func TestGetTopicReadAclIntegration(t *testing.T) {

	test_util.StartIntegrationTestWithContext(t, "sasl-admin")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	topicName := test_util.CreateTopic(t, "acl-topic")

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

	test_util.AssertContains(t, fmt.Sprintf("Topic %s Literal User:user * Read Allow", topicName), outputLines)
}
