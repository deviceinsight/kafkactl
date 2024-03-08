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

func TestDeleteSingleTopicIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName := testutil.CreateTopic(t, "topic-to-be-deleted")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("delete", "topic", topicName); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("topic deleted: %s", topicName), kafkaCtl.GetStdOut())

	verifyTopicDeleted(t, kafkaCtl, topicName)
}

func TestDeleteMultipleTopicsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	topicName1 := testutil.CreateTopic(t, "topic-to-be-deleted")
	topicName2 := testutil.CreateTopic(t, "topic-to-be-deleted")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("delete", "topic", topicName1, topicName2); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(outputLines) != 2 {
		t.Fatalf("unexpected output. expected two lines got %d: %s", len(outputLines), kafkaCtl.GetStdOut())
	}

	testutil.AssertEquals(t, fmt.Sprintf("topic deleted: %s", topicName1), outputLines[0])
	testutil.AssertEquals(t, fmt.Sprintf("topic deleted: %s", topicName2), outputLines[1])

	verifyTopicDeleted(t, kafkaCtl, topicName1)
	verifyTopicDeleted(t, kafkaCtl, topicName2)
}

func TestDeleteTopicAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "delete-complete-"

	topicName1 := testutil.CreateTopic(t, prefix+"a")
	topicName2 := testutil.CreateTopic(t, prefix+"b")
	topicName3 := testutil.CreateTopic(t, prefix+"c")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "delete", "topic", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, topicName1, outputLines)
	testutil.AssertContains(t, topicName2, outputLines)
	testutil.AssertContains(t, topicName3, outputLines)
}

func verifyTopicDeleted(t *testing.T, kafkaCtl testutil.KafkaCtlTestCommand, topicName string) {

	checkTopicDeleted := func(_ uint) error {
		_, err := kafkaCtl.Execute("get", "topics", "-o", "compact")

		if err != nil {
			return err
		}
		topics := strings.SplitN(kafkaCtl.GetStdOut(), "\n", -1)
		if util.ContainsString(topics, topicName) {
			return errors.New("topic not yet deleted")
		}
		return nil
	}

	err := retry.Retry(
		checkTopicDeleted,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("topic %s was not deleted: %v", topicName, err)
	}
}
