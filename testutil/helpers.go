package testutil

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/util"
	schemaregistry "github.com/landoop/schema-registry"
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

func CreateAvroTopic(t *testing.T, topicPrefix, keySchema, valueSchema string, flags ...string) string {

	topicName := CreateTopic(t, topicPrefix, flags...)

	schemaRegistry, err := schemaregistry.NewClient("localhost:18081")

	if err != nil {
		t.Fatalf("failed to create schema registry client: %v", err)
	}

	if keySchema != "" {
		if _, err := schemaRegistry.RegisterNewSchema(topicName+"-key", keySchema); err != nil {
			t.Fatalf("unable to register schema for key: %v", err)
		}
	}

	if valueSchema != "" {
		if _, err := schemaRegistry.RegisterNewSchema(topicName+"-value", valueSchema); err != nil {
			t.Fatalf("unable to register schema for value: %v", err)
		}
	}

	return topicName
}

func VerifyTopicExists(t *testing.T, topic string) {

	kafkaCtl := CreateKafkaCtlCommand()

	findTopic := func(attempt uint) error {
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

	findConsumerGroup := func(attempt uint) error {
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

	verifyConsumerOffset := func(attempt uint) error {
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

	verifyTopicNotInGroup := func(attempt uint) error {
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
