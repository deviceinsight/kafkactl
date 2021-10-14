package deletion_test

import (
	"context"
	"fmt"
	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/test_util"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/pkg/errors"
	"strings"
	"testing"
	"time"
)

func TestDeleteConsumerGroupOffsetIntegration(t *testing.T) {
	test_util.StartIntegrationTest(t)
	topicName := test_util.CreateTopic(t, "delete-consumer-group-offset", "--partitions", "3")

	client := test_util.CreateClient(t)
	defer client.Close()
	groupName := test_util.GetPrefixedName("cg-delete-offset-test")
	_, err := sarama.NewConsumerGroupFromClient(groupName, client)
	if err != nil {
		t.Fatalf("Fail to create consumer group %s", groupName)
	}

	// Create offset on the three partitions
	test_util.MarkOffset(t, client, groupName, topicName, 0, 0)
	test_util.MarkOffset(t, client, groupName, topicName, 1, 0)
	test_util.MarkOffset(t, client, groupName, topicName, 2, 0)

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	// Deleting an offset for a topic that does not exist shall fail
	if _, err := kafkaCtl.Execute("delete", "consumer-group-offset", groupName,
		"--topic", "this-topic-does-not-exist",
		"--partition", "0"); err == nil {
		t.Fatalf("Deleting an offset for a topic that does not exist shall fail")
	} else {
		test_util.AssertEquals(t, fmt.Sprintf("no offsets for topic: %s", "this-topic-does-not-exist"), err.Error())
	}

	// Deleting existing offset of a partition shall succeed
	// Offsets on other partitions must not be impacted
	if _, err := kafkaCtl.Execute("delete", "consumer-group-offset", groupName,
		"--topic", topicName,
		"--partition", "1"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	test_util.AssertEquals(t, offsetDeletedMessage(groupName, topicName, 1), kafkaCtl.GetStdOut())

	//verify that offset has been deleted for partition 1
	if err := checkOffsetDeleted(kafkaCtl, groupName, topicName, 1); err != nil {
		t.Fatal(err.Error())
	}
	//verify that offset still exists for partition 0 and 2
	if err := checkOffsetDeleted(kafkaCtl, groupName, topicName, 0); err == nil {
		t.Fatalf("offset for partition %d has been deleted", 0)
	}
	if err := checkOffsetDeleted(kafkaCtl, groupName, topicName, 2); err == nil {
		t.Fatalf("offset for partition %d has been deleted", 2)
	}

	// Deleting an offset that does not exist shall fail
	if _, err := kafkaCtl.Execute("delete", "consumer-group-offset", groupName,
		"--topic", topicName,
		"--partition", "1"); err == nil {
		t.Fatalf("Deleting an offset that does not exist shall fail")
	} else {
		test_util.AssertEquals(t, fmt.Sprintf("No offset for partition: %d", 1), err.Error())
	}

	// --partition=-1 shall delete all offsets on topic (here partitions 0 and 2)
	if _, err := kafkaCtl.Execute("delete", "consumer-group-offset", groupName,
		"--topic", topicName,
		"--partition", "-1"); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}
	//verify output messages
	messages := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
	if !util.ContainsString(messages, offsetDeletedMessage(groupName, topicName, 0)) {
		t.Fatalf("offset for partition %d not deleted", 0)
	}
	if !util.ContainsString(messages, offsetDeletedMessage(groupName, topicName, 2)) {
		t.Fatalf("offset for partition %d not deleted", 2)
	}
	//verify that offsets have been effectively deleted
	if err := checkOffsetDeleted(kafkaCtl, groupName, topicName, 0); err != nil {
		t.Fatal(err.Error())
	}
	if err := checkOffsetDeleted(kafkaCtl, groupName, topicName, 2); err != nil {
		t.Fatal(err.Error())
	}
}

func TestDeleteConsumerGroupOffsetOnActiveTopicIntegration(t *testing.T) {
	test_util.StartIntegrationTest(t)
	topicName := test_util.CreateTopic(t, "delete-consumer-group-offset-active", "--partitions", "1")

	client := test_util.CreateClient(t)
	defer client.Close()
	groupName := test_util.GetPrefixedName("cg-delete-offset-test")
	consumerGroup, err := sarama.NewConsumerGroupFromClient(groupName, client)
	if err != nil {
		t.Fatalf("Fail to create consumer group %s", groupName)
	}

	// Create offset on the three partitions
	test_util.MarkOffset(t, client, groupName, topicName, 0, 0)

	kafkaCtl := test_util.CreateKafkaCtlCommand()

	backgroundCtx := context.Background()

	consumer := consumerGrpHandler{
		ready: make(chan bool),
	}

	consumer.ready = make(chan bool)
	err = consumerGroup.Consume(backgroundCtx, []string{topicName}, &consumer)
	if err != nil {
		t.Fatal("fail to create consumer")
	}

	<-consumer.ready

	// Deleting an offset shall fail is consumer group is active on the topic
	if _, err := kafkaCtl.Execute("delete", "consumer-group-offset", groupName,
		"--topic", topicName,
		"--partition", "0"); err == nil {
		t.Fatalf("Deleting an offset shall fail is consumer group is active on the topic")
	} else {
		if !strings.HasPrefix(err.Error(), failedToDeleteMessage(groupName, topicName, 0)) {
			t.Fatalf("Unexpected error message: %s", err.Error())
		}
	}

	err = consumerGroup.Close()
	if err != nil {
		t.Fatal("fail to close consumer group")
	}

	// consumer group is closed, it shall be possible to delete offset
	if _, err := kafkaCtl.Execute("delete", "consumer-group-offset", groupName,
		"--topic", topicName,
		"--partition", "0"); err != nil {
		t.Fatalf("Deleting an offset failed")
	}
}

func TestDeleteConsumerGroupOffsetAutoCompletionIntegration(t *testing.T) {

	test_util.StartIntegrationTest(t)
	topicName := test_util.CreateTopic(t, "delete-consumer-group-completion")
	prefix := "delete-complete-"

	groupName1 := test_util.CreateConsumerGroup(t, topicName, prefix+"a")
	groupName2 := test_util.CreateConsumerGroup(t, topicName, prefix+"b")
	groupName3 := test_util.CreateConsumerGroup(t, topicName, prefix+"c")

	kafkaCtl := test_util.CreateKafkaCtlCommand()
	kafkaCtl.Verbose = false

	if _, err := kafkaCtl.Execute("__complete", "delete", "consumer-group-offset", ""); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	outputLines := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")

	test_util.AssertContains(t, groupName1, outputLines)
	test_util.AssertContains(t, groupName2, outputLines)
	test_util.AssertContains(t, groupName3, outputLines)

}

func offsetDeletedMessage(groupName string, topic string, partition int32) string {
	return fmt.Sprintf("consumer-group-offset deleted: [group: %s, topic: %s, partition: %d]",
		groupName, topic, partition)
}
func failedToDeleteMessage(groupName string, topic string, partition int32) string {
	return fmt.Sprintf("failed to delete consumer-group-offset [group: %s, topic: %s, partition: %d]",
		groupName, topic, partition)
}

func checkOffsetDeleted(kafkaCtl test_util.KafkaCtlTestCommand, groupName string, topic string, partition int32) error {
	checkOffsetDeleted := func(attempt uint) error {
		_, err := kafkaCtl.Execute("describe", "consumer-group", groupName, "-o", "yaml")

		if err != nil {
			return err
		} else {
			consumerGroupDescStr := strings.Split(strings.TrimSpace(kafkaCtl.GetStdOut()), "\n")
			if util.ContainsString(consumerGroupDescStr, fmt.Sprintf("  - partition: %d", partition)) {
				return errors.New("consumer-group-offset not exists")
			} else {
				return nil
			}
		}
	}

	err := retry.Retry(
		checkOffsetDeleted,
		strategy.Limit(5),
		strategy.Backoff(backoff.Linear(10*time.Millisecond)),
	)

	if err != nil {
		return errors.Wrapf(err, "consumer-group-offset [%s, %s, %d] exists: %v", groupName, topic, partition, err)
	} else {
		return nil
	}
}

type consumerGrpHandler struct {
	ready chan bool
}

func (h consumerGrpHandler) Setup(_ sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(h.ready)
	return nil
}
func (consumerGrpHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerGrpHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	return nil
}
