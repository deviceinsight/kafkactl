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

func TestAlterPartitionAutoCompletionIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	prefix := "alter-p-complete-"

	topicName1 := test_util.CreateTopic(t, prefix+"a", "--partitions", "2")
	topicName2 := test_util.CreateTopic(t, prefix+"b")
	topicName3 := test_util.CreateTopic(t, prefix+"c")

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "alter", "partition", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, topicName1, outputLines)
	test_util.AssertContains(t, topicName2, outputLines)
	test_util.AssertContains(t, topicName3, outputLines)

	if _, err := kafkaCtl.Execute("__complete", "alter", "partition", topicName1, ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines = strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, "0", outputLines)
	test_util.AssertContains(t, "1", outputLines)
}

func TestAlterPartitionReplicationFactorIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)

	prefix := "alter-p-replicas-"

	topicName := test_util.CreateTopic(t, prefix, "--partitions", "2", "--replication-factor", "3")

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "partition", topicName, "0", "--replicas", "101,102"); err != nil {
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

			if len(topic.Partitions) != 2 {
				return errors.Newf("expected 2 partitions, but was %d", len(topic.Partitions))
			}

			if len(topic.Partitions[0].Replicas) == 2 && len(topic.Partitions[1].Replicas) == 3 {
				if topic.Partitions[0].Replicas[0] == 101 && topic.Partitions[0].Replicas[1] == 102 {
					return nil
				} else {
					return errors.Newf("different brokers expected %v", topic.Partitions[0].Replicas)
				}
			} else {
				return errors.Newf("replica count incorrect %v", topic.Partitions)
			}
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
