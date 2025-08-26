package describe_test

import (
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/topic"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestDescribeTopicConfigsIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "describe-t-configs-"

	topicName1 := testutil.CreateTopic(t, prefix, "--config", "retention.ms=3600000")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	// default --print-configs=no_defaults
	if _, err := kafkaCtl.Execute("describe", "topic", topicName1, "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	describedTopic, err := topic.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}

	configKeys := getConfigKeys(describedTopic.Configs)

	testutil.AssertArraysEquals(t, []string{"retention.ms"}, configKeys)
	testutil.AssertEquals(t, "3600000", describedTopic.Configs[0].Value)

	if _, err := kafkaCtl.Execute("describe", "topic", topicName1, "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	describedTopic, err = topic.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}

	configKeys = getConfigKeys(describedTopic.Configs)

	testutil.AssertArraysEquals(t, []string{"retention.ms"}, configKeys)
	testutil.AssertEquals(t, "3600000", describedTopic.Configs[0].Value)

	// all configs
	if _, err := kafkaCtl.Execute("describe", "topic", topicName1, "--all-configs", "-o", "yaml"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	describedTopic, err = topic.FromYaml(kafkaCtl.GetStdOut())
	if err != nil {
		t.Fatalf("failed to read yaml: %v", err)
	}

	configKeys = getConfigKeys(describedTopic.Configs)

	testutil.AssertContains(t, "retention.ms", configKeys)
	testutil.AssertContains(t, "cleanup.policy", configKeys)
}

func TestDescribeTopicAutoCompletionIntegration(t *testing.T) {

	testutil.StartIntegrationTest(t)

	prefix := "describe-t-complete-"

	topicName1 := testutil.CreateTopic(t, prefix+"a")
	topicName2 := testutil.CreateTopic(t, prefix+"b")
	topicName3 := testutil.CreateTopic(t, prefix+"c")

	kafkaCtl := testutil.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "describe", "topic", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	testutil.AssertContains(t, topicName1, outputLines)
	testutil.AssertContains(t, topicName2, outputLines)
	testutil.AssertContains(t, topicName3, outputLines)
}

func getConfigKeys(configs []internal.Config) []string {
	keys := make([]string, len(configs))
	for i, config := range configs {
		keys[i] = config.Name
	}
	return keys
}
