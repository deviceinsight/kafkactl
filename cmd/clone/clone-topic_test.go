package clone_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"gopkg.in/errgo.v2/fmt/errors"

	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/deviceinsight/kafkactl/testutil"
)

func TestCloneTopicIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	srcTopic := testutil.CreateTopic(t, "src-topic", "-p", "2", "-r", "3", "-c", "retention.ms=86400000")
	targetTopic := testutil.GetPrefixedName("target-topic")

	if _, err := kafkaCtl.Execute("clone", "topic", srcTopic, targetTopic); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	testutil.AssertEquals(t, fmt.Sprintf("topic %s cloned to %s", srcTopic, targetTopic), kafkaCtl.GetStdOut())

	getTopic := func(attempt uint) error {
		_, err := kafkaCtl.Execute("describe", "topic", targetTopic, "-o", "yaml")

		if err != nil {
			return err
		}
		topic, err := topic.FromYaml(kafkaCtl.GetStdOut())
		if err != nil {
			return err
		}
		if len(topic.Partitions) != 2 {
			return errors.Newf("only the following partitions present: %v", topic.Partitions)
		}

		if len(topic.Partitions[0].Replicas) != 3 || len(topic.Partitions[1].Replicas) != 3 {
			return errors.Newf("replication factor is not 3: %v", topic.Partitions)
		}

		var retention string
		for _, configEntry := range topic.Configs {
			if configEntry.Name == "retention.ms" {
				retention = configEntry.Value
				break
			}
		}

		if retention != "86400000" {
			return errors.Newf("topic retention is not 86400000: %v", retention)
		}

		return nil
	}

	err := retry.Retry(
		getTopic,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("could not get topic %s: %v", targetTopic, err)
	}
}

func TestCloneNonExistingTopicIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	srcTopic := testutil.GetPrefixedName("src-topic")
	targetTopic := testutil.GetPrefixedName("target-topic")

	if _, err := kafkaCtl.Execute("clone", "topic", srcTopic, targetTopic); err != nil {
		testutil.AssertErrorContains(t, fmt.Sprintf("topic '%s' does not exist", srcTopic), err)
	} else {
		t.Fatalf("Expected clone operation to fail")
	}
}

func TestCloneToExistingTopicIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	srcTopic := testutil.CreateTopic(t, "src-topic")
	targetTopic := testutil.CreateTopic(t, "target-topic")

	if _, err := kafkaCtl.Execute("clone", "topic", srcTopic, targetTopic); err != nil {
		testutil.AssertErrorContains(t, fmt.Sprintf("topic '%s' already exists", targetTopic), err)
	} else {
		t.Fatalf("Expected clone operation to fail")
	}
}
