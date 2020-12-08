package alter_test

import (
	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/test_util"
	"gopkg.in/errgo.v2/fmt/errors"
	"strings"
	"testing"
	"time"
)

func TestAlterTopicAutoCompletionIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	prefix := "alter-t-complete-"

	topicName1 := test_util.CreateTopic(t, prefix+"a")
	topicName2 := test_util.CreateTopic(t, prefix+"b")
	topicName3 := test_util.CreateTopic(t, prefix+"c")

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "alter", "topic", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, topicName1, outputLines)
	test_util.AssertContains(t, topicName2, outputLines)
	test_util.AssertContains(t, topicName3, outputLines)
}

func TestAlterTopicPartitionsIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	prefix := "alter-t-partition-"

	topicName := test_util.CreateTopic(t, prefix)

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "topic", topicName, "--partitions", "32"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	getPartitions := func(attempt uint) error {
		_, err := kafkaCtl.Execute("describe", "topic", topicName, "-o", "yaml")

		if err != nil {
			return err
		} else {
			topic, err := operations.TopicFromYaml(kafkaCtl.GetStdOut())
			if err != nil {
				return err
			}

			if len(topic.Partitions) == 32 {
				return nil
			} else {
				return errors.Newf("only the following partitions present: %v", topic.Partitions)
			}
		}
	}

	err := retry.Retry(
		getPartitions,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("could not get partitions %s: %v", topicName, err)
	}
}

func TestAlterTopicIncreaseReplicationFactorIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	prefix := "alter-t-ireplicas-"

	topicName := test_util.CreateTopic(t, prefix, "--partitions", "32")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "topic", topicName, "--replication-factor", "3"); err != nil {
		test_util.AssertErrorContains(t, "version of API is not supported", err)
		return
	}

	checkReplicas := func(attempt uint) error {
		_, err := kafkaCtl.Execute("describe", "topic", topicName, "-o", "yaml")

		if err != nil {
			return err
		} else {
			topic, err := operations.TopicFromYaml(kafkaCtl.GetStdOut())
			if err != nil {
				return err
			}

			for _, p := range topic.Partitions {
				if len(p.Replicas) != 3 {
					return errors.Newf("partition %d has %d replicas != 3", p.Id, len(p.Replicas))
				}
			}

			return nil
		}
	}

	err := retry.Retry(
		checkReplicas,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("could not check Replicas for topic %s: %v", topicName, err)
	}
}

func TestAlterTopicDecreaseReplicationFactorIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	prefix := "alter-t-dreplicas-"

	topicName := test_util.CreateTopic(t, prefix, "--partitions", "32", "--replication-factor", "3")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "topic", topicName, "--replication-factor", "1"); err != nil {
		test_util.AssertErrorContains(t, "version of API is not supported", err)
		return
	}

	checkReplicas := func(attempt uint) error {
		_, err := kafkaCtl.Execute("describe", "topic", topicName, "-o", "yaml")

		if err != nil {
			return err
		} else {
			topic, err := operations.TopicFromYaml(kafkaCtl.GetStdOut())
			if err != nil {
				return err
			}

			for _, p := range topic.Partitions {
				if len(p.Replicas) != 1 {
					return errors.Newf("partition %d has %d replicas != 1", p.Id, len(p.Replicas))
				}
			}

			return nil
		}
	}

	err := retry.Retry(
		checkReplicas,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("could not check Replicas for topic %s: %v", topicName, err)
	}
}
