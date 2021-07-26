package deletion_test

import (
	"errors"
	"fmt"
	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/test_util"
	"github.com/deviceinsight/kafkactl/util"
	"strings"
	"testing"
	"time"
)

func TestDeleteSingleConsumerGroupIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)
	topicName := test_util.CreateTopic(t, "delete-consumer-group-single")
	groupName := test_util.CreateConsumerGroup(t, topicName, "consumer-group-to-be-deleted")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("delete", "consumer-group", groupName); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, fmt.Sprintf("consumer-group deleted: %s", groupName), kafkaCtl.GetStdOut())

	verifyConsumerGroupDeleted(t, kafkaCtl, groupName)
}

func TestDeleteMultipleConsumerGroupsIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)
	topicName := test_util.CreateTopic(t, "delete-consumer-group-multiple")
	groupName1 := test_util.CreateConsumerGroup(t, topicName, "consumer-group-to-be-deleted")
	groupName2 := test_util.CreateConsumerGroup(t, topicName, "consumer-group-to-be-deleted")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("delete", "consumer-group", groupName1, groupName2); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(outputLines) != 2 {
		t.Fatalf("unexpected output. expected two lines got %d: %s", len(outputLines), kafkaCtl.GetStdOut())
	}

	test_util.AssertEquals(t, fmt.Sprintf("consumer-group deleted: %s", groupName1), outputLines[0])
	test_util.AssertEquals(t, fmt.Sprintf("consumer-group deleted: %s", groupName2), outputLines[1])

	verifyConsumerGroupDeleted(t, kafkaCtl, groupName1)
	verifyConsumerGroupDeleted(t, kafkaCtl, groupName2)
}

func TestDeleteConsumerGroupAutoCompletionIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)
	topicName := test_util.CreateTopic(t, "delete-consumer-group-completion")
	prefix := "delete-complete-"

	groupName1 := test_util.CreateConsumerGroup(t, topicName, prefix+"a")
	groupName2 := test_util.CreateConsumerGroup(t, topicName, prefix+"b")
	groupName3 := test_util.CreateConsumerGroup(t, topicName, prefix+"c")

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "delete", "consumer-group", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, groupName1, outputLines)
	test_util.AssertContains(t, groupName2, outputLines)
	test_util.AssertContains(t, groupName3, outputLines)
}

func verifyConsumerGroupDeleted(t *testing.T, kafkaCtl test_util.KafkaCtlTestCommand, groupName string) {

	checkConsumerGrouDeleted := func(attempt uint) error {
		_, err := kafkaCtl.Execute("get", "consumer-groups", "-o", "compact")

		if err != nil {
			return err
		} else {
			consumerGroups := strings.SplitN(kafkaCtl.GetStdOut(), "\n", -1)
			if util.ContainsString(consumerGroups, groupName) {
				return errors.New("consumer-group not yet deleted")
			} else {
				return nil
			}
		}
	}

	err := retry.Retry(
		checkConsumerGrouDeleted,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("consumer-group %s was not deleted: %v", groupName, err)
	}
}
