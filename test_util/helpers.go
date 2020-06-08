package test_util

import (
	"errors"
	"fmt"
	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/util"
	"strings"
	"testing"
	"time"
)

func CreateTopic(t *testing.T, topicPrefix string, flags ...string) string {

	kafkaCtl := CreateKafkaCtlCommand()
	topicName := GetTopicName(topicPrefix)

	createTopicWithFlags := append([]string{"create", "topic", topicName}, flags...)

	if _, err := kafkaCtl.Execute(createTopicWithFlags...); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	AssertEquals(t, fmt.Sprintf("topic created: %s", topicName), kafkaCtl.GetStdOut())

	VerifyTopicExists(t, topicName)

	return topicName
}

func VerifyTopicExists(t *testing.T, topic string) {

	kafkaCtl := CreateKafkaCtlCommand()

	findTopic := func(attempt uint) error {
		_, err := kafkaCtl.Execute("get", "topics", "-o", "compact")

		if err != nil {
			return err
		} else {
			topics := strings.SplitN(kafkaCtl.GetStdOut(), "\n", -1)
			if util.ContainsString(topics, topic) {
				return nil
			} else {
				return errors.New("topic not in list")
			}
		}
	}

	err := retry.Retry(
		findTopic,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("could not find topic %s: %v", topic, err)
	}
}
