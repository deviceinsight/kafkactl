package testutil

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/deviceinsight/kafkactl/v5/internal/output"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/v5/internal/util"
	"github.com/riferrei/srclient"
)

func CreateTopic(t *testing.T, topicPrefix string, flags ...string) string {
	kafkaCtl := CreateKafkaCtlCommand()
	topicName := GetPrefixedName(topicPrefix)

	createTopicWithFlags := append([]string{"create", "topic", topicName}, flags...)

	if _, err := kafkaCtl.Execute(createTopicWithFlags...); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	AssertEquals(t, fmt.Sprintf("topic created: %s", topicName), kafkaCtl.GetStdOut())

	VerifyTopicExists(t, topicName)

	return topicName
}

func RegisterSchema(t *testing.T, subjectName, schema string, schemaType srclient.SchemaType, references ...srclient.Reference) {
	schemaRegistry := srclient.NewSchemaRegistryClient("http://localhost:18081")

	if schema, err := schemaRegistry.CreateSchema(subjectName, schema, schemaType, references...); err != nil {
		t.Fatalf("unable to register schema for value: %v", err)
	} else {
		output.TestLogf("registered schema %q with ID=%d", schema, schema.ID())
	}
}

func CreateTopicWithSchema(t *testing.T, topicPrefix, keySchema, valueSchema string, schemaType srclient.SchemaType,
	flags ...string,
) string {
	topicName := CreateTopic(t, topicPrefix, flags...)

	schemaRegistry := srclient.NewSchemaRegistryClient("http://localhost:18081")

	if keySchema != "" {
		if schema, err := schemaRegistry.CreateSchema(topicName+"-key", keySchema, schemaType); err != nil {
			t.Fatalf("unable to register schema for key: %v", err)
		} else {
			output.TestLogf("registered schema %q with ID=%d", topicName+"-key", schema.ID())
		}
	}

	if valueSchema != "" {
		if schema, err := schemaRegistry.CreateSchema(topicName+"-value", valueSchema, schemaType); err != nil {
			t.Fatalf("unable to register schema for value: %v", err)
		} else {
			output.TestLogf("registered schema %q with ID=%d", topicName+"-value", schema.ID())
		}
	}

	return topicName
}

func VerifyTopicExists(t *testing.T, topic string) {
	kafkaCtl := CreateKafkaCtlCommand()

	findTopic := func(_ uint) error {
		_, err := kafkaCtl.Execute("get", "topics", "-o", "compact")
		if err != nil {
			return err
		}
		topics := strings.SplitN(kafkaCtl.GetStdOut(), "\n", -1)
		if util.ContainsString(topics, topic) {
			return nil
		}
		return errors.New("topic not in list")
	}

	err := retry.Retry(
		findTopic,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)
	if err != nil {
		t.Fatalf("could not find topic %s: %v", topic, err)
	}

	// add a sleep here, so that the new topic is known by all
	// brokers hopefully
	time.Sleep(1000 * time.Millisecond)
}

func CreateConsumerGroup(t *testing.T, groupPrefix string, topics ...string) string {
	kafkaCtl := CreateKafkaCtlCommand()
	groupName := GetPrefixedName(groupPrefix)

	createCgo := []string{"create", "consumer-group", groupName, "--newest"}

	for _, topic := range topics {
		createCgo = append(createCgo, "--topic", topic)
	}

	if _, err := kafkaCtl.Execute(createCgo...); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	AssertEquals(t, fmt.Sprintf("consumer-group created: %s", groupName), kafkaCtl.GetStdOut())

	VerifyGroupExists(t, groupName)

	return groupName
}

func ProduceMessage(t *testing.T, topic, key, value string, expectedPartition, expectedOffset int64) {
	kafkaCtl := CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topic, "--key", key, "--value", value); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	AssertEquals(t, fmt.Sprintf("message produced (partition=%d\toffset=%d)", expectedPartition, expectedOffset), kafkaCtl.GetStdOut())
}

func ProduceMessageOnPartition(t *testing.T, topic, key, value string, partition int32, expectedOffset int64) {
	kafkaCtl := CreateKafkaCtlCommand()

	if _, err := kafkaCtl.Execute("produce", topic, "--key", key, "--value", value, "--partition", strconv.FormatInt(int64(partition), 10)); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	AssertEquals(t, fmt.Sprintf("message produced (partition=%d\toffset=%d)", partition, expectedOffset), kafkaCtl.GetStdOut())
}

func VerifyGroupExists(t *testing.T, group string) {
	kafkaCtl := CreateKafkaCtlCommand()

	findConsumerGroup := func(_ uint) error {
		_, err := kafkaCtl.Execute("get", "cg", "-o", "compact")
		if err != nil {
			return err
		}
		groups := strings.SplitN(kafkaCtl.GetStdOut(), "\n", -1)
		if util.ContainsString(groups, group) {
			return nil
		}
		return errors.New("group not in list")
	}

	err := retry.Retry(
		findConsumerGroup,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)
	if err != nil {
		t.Fatalf("could not find group %s: %v", group, err)
	}

	// add a sleep here, so that the new group is known by all
	// brokers hopefully
	time.Sleep(500 * time.Millisecond)
}

func VerifyConsumerGroupOffset(t *testing.T, group, topic string, expectedConsumerOffset int) {
	kafkaCtl := CreateKafkaCtlCommand()

	consumerOffsetRegex, _ := regexp.Compile(`consumerOffset: (\d)`)

	verifyConsumerOffset := func(_ uint) error {
		_, err := kafkaCtl.Execute("describe", "cg", group, "--topic", topic, "-o", "yaml")
		if err != nil {
			return err
		}

		match := consumerOffsetRegex.FindStringSubmatch(kafkaCtl.GetStdOut())

		if len(match) == 2 {
			if match[1] == fmt.Sprintf("%d", expectedConsumerOffset) {
				return nil
			}
			return fmt.Errorf("unexpected consumer offset %s != %d", match[1], expectedConsumerOffset)
		}

		return errors.New("cannot find consumer offset")
	}

	err := retry.Retry(
		verifyConsumerOffset,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)
	if err != nil {
		t.Fatalf("failed to verify offset for group=%s topic=%s: %v", group, topic, err)
	}
}

func VerifyTopicNotInConsumerGroup(t *testing.T, group, topic string) {
	kafkaCtl := CreateKafkaCtlCommand()

	emptyTopicsRegex, _ := regexp.Compile(`topics: \[]`)

	verifyTopicNotInGroup := func(_ uint) error {
		_, err := kafkaCtl.Execute("describe", "cg", group, "--topic", topic, "-o", "yaml")
		if err != nil {
			return err
		}

		if emptyTopicsRegex.MatchString(kafkaCtl.GetStdOut()) {
			return nil
		}
		return errors.New("expected topic not to be part of consumer-group")
	}

	err := retry.Retry(
		verifyTopicNotInGroup,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)
	if err != nil {
		t.Fatalf("failed to verify topic=%s not in group=%s: %v", topic, group, err)
	}
}
