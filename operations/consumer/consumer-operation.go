package consumer

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ConsumerFlags struct {
	PrintKeys       bool
	PrintTimestamps bool
	PrintAvroSchema bool
	OutputFormat    string
	Partitions      []int
	Offsets         []string
	FromBeginning   bool
	BufferSize      int
	EncodeValue     string
	EncodeKey       string
}

type offset struct {
	relative bool
	start    int64
	diff     int64
}

type Interval struct {
	start offset
	end   offset
}

type ConsumedMessage struct {
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Timestamp *time.Time
}

type ConsumerOperation struct {
}

func (operation *ConsumerOperation) Consume(topic string, flags ConsumerFlags) {

	clientContext := operations.CreateClientContext()

	var (
		err       error
		client    sarama.Client
		topExists bool
	)

	if client, err = operations.CreateClient(&clientContext); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if topExists, err = operations.TopicExists(&client, topic); err != nil {
		output.Failf("failed to read topics err=%v", err)
	}

	if !topExists {
		output.Failf("topic '%s' does not exist", topic)
	}

	c, err := sarama.NewConsumer(clientContext.Brokers, nil)
	if err != nil {
		output.Failf("Failed to start consumer: %s", err)
	}

	var deserializer MessageDeserializer

	if clientContext.AvroSchemaRegistry != "" {
		deserializer = CreateAvroMessageDeserializer(topic, clientContext.AvroSchemaRegistry)
	} else {
		deserializer = DefaultMessageDeserializer{}
	}

	var partitions []int32

	if flags.Partitions == nil || len(flags.Partitions) == 0 {
		partitions, err = c.Partitions(topic)

		if err != nil {
			output.Failf("Failed to get the list of partitions: %s", err)
		}
	} else {
		for partition := range flags.Partitions {
			partitions = append(partitions, int32(partition))
		}
	}

	var (
		messages = make(chan *sarama.ConsumerMessage, flags.BufferSize)
		closing  = make(chan struct{})
		wg       sync.WaitGroup
	)

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Kill, os.Interrupt)
		<-signals
		output.Debugf("Initiating shutdown of consumer...")
		close(closing)
	}()

	output.Debugf("Start consuming topic: %s", topic)

	for _, partition := range partitions {
		initialOffset := getInitialOffset(flags, partition)
		pc, err := c.ConsumePartition(topic, partition, initialOffset)
		if err != nil {
			output.Failf("Failed to start consumer for partition %d: %s", partition, err)
		}

		output.Debugf("Start consuming partition %d from offset %d", partition, initialOffset)

		go func(pc sarama.PartitionConsumer) {
			<-closing
			pc.AsyncClose()
		}(pc)

		wg.Add(1)
		go func(pc sarama.PartitionConsumer) {
			defer wg.Done()
			for message := range pc.Messages() {
				messages <- message
			}
		}(pc)
	}

	go func() {
		for msg := range messages {
			deserializer.Deserialize(msg, flags)
		}
	}()

	wg.Wait()
	output.Debugf("Done consuming topic: %s", topic)
	close(messages)

	if err := c.Close(); err != nil {
		output.Failf("Failed to close consumer: ", err)
	}
}

func getInitialOffset(flags ConsumerFlags, currentPartition int32) int64 {

	for _, offsetFlag := range flags.Offsets {
		offsetParts := strings.Split(offsetFlag, "=")

		if len(offsetParts) == 2 {

			partition, err := strconv.Atoi(offsetParts[0])
			if err != nil {
				output.Failf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
			}

			if int32(partition) != currentPartition {
				continue
			}

			offset, err := strconv.ParseInt(offsetParts[1], 10, 64)
			if err != nil {
				output.Failf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
			}

			return offset
		}
	}

	if flags.FromBeginning {
		return sarama.OffsetOldest
	} else {
		return sarama.OffsetNewest
	}
}
