package test_util

import (
	"errors"
	"fmt"
	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/deviceinsight/kafkactl/util"
	schemaregistry "github.com/landoop/schema-registry"
	"strings"
	"testing"
	"time"
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

	// add a sleep here, so that the new topic is known by all
	// brokers hopefully
	time.Sleep(200 * time.Millisecond)
}

func CreateConsumerGroup(t *testing.T, topic string, groupPrefix string) string {

	kafkaCtl := CreateKafkaCtlCommand()
	groupName := GetPrefixedName(groupPrefix)

	createCgo := []string{"create", "consumer-group", groupName, "--topic", topic, "--newest"}

	if _, err := kafkaCtl.Execute(createCgo...); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	AssertEquals(t, fmt.Sprintf("consumer-group created: %s", groupName), kafkaCtl.GetStdOut())

	VerifyGroupExists(t, groupName)

	return groupName
}

func VerifyGroupExists(t *testing.T, group string) {

	kafkaCtl := CreateKafkaCtlCommand()

	findTopic := func(attempt uint) error {
		_, err := kafkaCtl.Execute("get", "cg", "-o", "compact")

		if err != nil {
			return err
		} else {
			groups := strings.SplitN(kafkaCtl.GetStdOut(), "\n", -1)
			if util.ContainsString(groups, group) {
				return nil
			} else {
				return errors.New("group not in list")
			}
		}
	}

	err := retry.Retry(
		findTopic,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		t.Fatalf("could not find group %s: %v", group, err)
	}

	// add a sleep here, so that the new group is known by all
	// brokers hopefully
	time.Sleep(200 * time.Millisecond)
}
