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

func TestDeleteSingleTopicIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName := test_util.CreateTopic(t, "topic-to-be-deleted")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("delete", "topic", topicName); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, fmt.Sprintf("topic deleted: %s", topicName), kafkaCtl.GetStdOut())

	verifyTopicDeleted(t, kafkaCtl, topicName)
}

func TestDeleteMultipleTopicsIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	topicName1 := test_util.CreateTopic(t, "topic-to-be-deleted")
	topicName2 := test_util.CreateTopic(t, "topic-to-be-deleted")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("delete", "topic", topicName1, topicName2); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	if len(outputLines) != 2 {
		t.Fatalf("unexpected output. expected two lines got %d: %s", len(outputLines), kafkaCtl.GetStdOut())
	}

	test_util.AssertEquals(t, fmt.Sprintf("topic deleted: %s", topicName1), outputLines[0])
	test_util.AssertEquals(t, fmt.Sprintf("topic deleted: %s", topicName2), outputLines[1])

	verifyTopicDeleted(t, kafkaCtl, topicName1)
	verifyTopicDeleted(t, kafkaCtl, topicName2)
}

func verifyTopicDeleted(t *testing.T, kafkaCtl test_util.KafkaCtlTestCommand, topicName string) {

	checkTopicDeleted := func(attempt uint) error {
		_, err := kafkaCtl.Execute("get", "topics", "-o", "compact")

		if err != nil {
			return err
		} else {
			topics := strings.SplitN(kafkaCtl.GetStdOut(), "\n", -1)
			if util.ContainsString(topics, topicName) {
				return errors.New("topic not yet deleted")
			} else {
				return nil
			}
		}
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
