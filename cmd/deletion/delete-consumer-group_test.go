package deletion_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/internal/testutil"
	"github.com/deviceinsight/kafkactl/internal/util"
)

func TestDeleteSingleConsumerGroupIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)
	topicName := testutil.CreateTopic(t, "delete-consumer-group-single")
	groupName := testutil.CreateConsumerGroup(t, "consumer-group-to-be-deleted", topicName)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("delete", "consumer-group", groupName); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("consumer-group deleted: %s", groupName), kafkaCtl.GetStdOut())

	verifyConsumerGroupDeleted(t, kafkaCtl, groupName)
}

func TestDeleteMultipleConsumerGroupsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)
	topicName := testutil.CreateTopic(t, "delete-consumer-group-multiple")
	groupName1 := testutil.CreateConsumerGroup(t, "consumer-group-to-be-deleted", topicName)
	groupName2 := testutil.CreateConsumerGroup(t, "consumer-group-to-be-deleted", topicName)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("delete", "consumer-group", groupName1, groupName2); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(outputLines) != 2 {
		t.Fatalf("unexpected output. expected two lines got %d: %s", len(outputLines), kafkaCtl.GetStdOut())
	}

	testutil.AssertEquals(t, fmt.Sprintf("consumer-group deleted: %s", groupName1), outputLines[0])
	testutil.AssertEquals(t, fmt.Sprintf("consumer-group deleted: %s", groupName2), outputLines[1])

	verifyConsumerGroupDeleted(t, kafkaCtl, groupName1)
	verifyConsumerGroupDeleted(t, kafkaCtl, groupName2)
}

func TestDeleteConsumerGroupAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)
	topicName := testutil.CreateTopic(t, "delete-consumer-group-completion")
	prefix := "delete-complete-"

	groupName1 := testutil.CreateConsumerGroup(t, prefix+"a", topicName)
	groupName2 := testutil.CreateConsumerGroup(t, prefix+"b", topicName)
	groupName3 := testutil.CreateConsumerGroup(t, prefix+"c", topicName)

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "delete", "consumer-group", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, groupName1, outputLines)
	testutil.AssertContains(t, groupName2, outputLines)
	testutil.AssertContains(t, groupName3, outputLines)
}

func verifyConsumerGroupDeleted(t *testing.T, kafkaCtl testutil.KafkaCtlTestCommand, groupName string) {

	checkConsumerGrouDeleted := func(_ uint) error {
		_, err := kafkaCtl.Execute("get", "consumer-groups", "-o", "compact")

		if err != nil {
			return err
		}
		consumerGroups := strings.SplitN(kafkaCtl.GetStdOut(), "\n", -1)
		if util.ContainsString(consumerGroups, groupName) {
			return errors.New("consumer-group not yet deleted")
		}
		return nil
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
