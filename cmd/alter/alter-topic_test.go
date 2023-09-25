package alter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/internal/topic"
	"github.com/deviceinsight/kafkactl/testutil"
	"gopkg.in/errgo.v2/fmt/errors"
)

func TestAlterTopicAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "alter-t-complete-"

	topicName1 := testutil.CreateTopic(t, prefix+"a")
	topicName2 := testutil.CreateTopic(t, prefix+"b")
	topicName3 := testutil.CreateTopic(t, prefix+"c")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "alter", "topic", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, topicName1, outputLines)
	testutil.AssertContains(t, topicName2, outputLines)
	testutil.AssertContains(t, topicName3, outputLines)
}

func TestAlterTopicPartitionsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "alter-t-partition-"

	topicName := testutil.CreateTopic(t, prefix)

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "topic", topicName, "--partitions", "32"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	getPartitions := func(attempt uint) error {
		_, err := kafkaCtl.Execute("describe", "topic", topicName, "-o", "yaml")

		if err != nil {
			return err
		}
		topic, err := topic.FromYaml(kafkaCtl.GetStdOut())
		if err != nil {
			return err
		}
		if len(topic.Partitions) == 32 {
			return nil
		}
		return errors.Newf("only the following partitions present: %v", topic.Partitions)
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

	testutil.StartIntegrationTest(t)

	prefix := "alter-t-ireplicas-"

	topicName := testutil.CreateTopic(t, prefix, "--partitions", "32")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "topic", topicName, "--replication-factor", "3"); err != nil {
		testutil.AssertErrorContains(t, "version of API is not supported", err)
		return
	}

	checkReplicas := func(attempt uint) error {
		_, err := kafkaCtl.Execute("describe", "topic", topicName, "-o", "yaml")

		if err != nil {
			return err
		}
		topic, err := topic.FromYaml(kafkaCtl.GetStdOut())
		if err != nil {
			return err
		}
		for _, p := range topic.Partitions {
			if len(p.Replicas) != 3 {
				return errors.Newf("partition %d has %d replicas != 3", p.ID, len(p.Replicas))
			}
		}
		return nil
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

	testutil.StartIntegrationTest(t)

	prefix := "alter-t-dreplicas-"

	topicName := testutil.CreateTopic(t, prefix, "--partitions", "32", "--replication-factor", "3")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("alter", "topic", topicName, "--replication-factor", "1"); err != nil {
		testutil.AssertErrorContains(t, "version of API is not supported", err)
		return
	}

	checkReplicas := func(attempt uint) error {
		_, err := kafkaCtl.Execute("describe", "topic", topicName, "-o", "yaml")

		if err != nil {
			return err
		}
		topic, err := topic.FromYaml(kafkaCtl.GetStdOut())
		if err != nil {
			return err
		}
		for _, p := range topic.Partitions {
			if len(p.Replicas) != 1 {
				return errors.Newf("partition %d has %d replicas != 1", p.ID, len(p.Replicas))
			}
		}
		return nil
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

func TestAlterTopicConfigK8sIntegration(t *testing.T) {

	testutil.StartIntegrationTestWithContext(t, "k8s-mock")

	kafkaCtl := testutil.CreateKafkaCtlCommand()

	type testCases struct {
		description      string
		args             []string
		wantInKubectlCmd []string
	}

	for _, test := range []testCases{
		{
			description:      "single_config_defined_with_space",
			args:             []string{"alter", "topic", "fake-topic", "--config", "retention.ms=86400000"},
			wantInKubectlCmd: []string{"--config=retention.ms=86400000"},
		},
		{
			description:      "single_config_defined_with_equal",
			args:             []string{"alter", "topic", "fake-topic", "--config=retention.ms=86400000"},
			wantInKubectlCmd: []string{"--config=retention.ms=86400000"},
		},
		{
			description: "multiple_configs",
			args: []string{"alter", "topic", "fake-topic", "--config", "retention.ms=86400000",
				"--config", "cleanup.policy=compact"},
			wantInKubectlCmd: []string{"--config=retention.ms=86400000", "--config=cleanup.policy=compact"},
		},
	} {
		t.Run(test.description, func(t *testing.T) {

			if _, err := kafkaCtl.Execute(test.args...); err != nil {
				t.Fatalf("failed to execute command: %v", err)
			}

			output := kafkaCtl.GetStdOut()

			for _, wanted := range test.wantInKubectlCmd {
				testutil.AssertContainSubstring(t, wanted, output)
			}
		})
	}
}
