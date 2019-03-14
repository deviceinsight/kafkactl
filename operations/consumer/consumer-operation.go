package consumer

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"os"
	"os/signal"
	"sync"
	"time"
)

type ConsumerFlags struct {
	PrintKeys       bool
	PrintTimestamps bool
	PrintAvroSchema bool
	OutputFormat    string
	Partitions      []int
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

	var initialOffset int64
	if flags.FromBeginning {
		initialOffset = sarama.OffsetOldest
	} else {
		initialOffset = sarama.OffsetNewest
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
		output.Infof("Initiating shutdown of consumer...")
		close(closing)
	}()

	for _, partition := range partitions {
		pc, err := c.ConsumePartition(topic, partition, initialOffset)
		if err != nil {
			output.Failf("Failed to start consumer for partition %d: %s", partition, err)
		}

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
	output.Infof("Done consuming topic: %s", topic)
	close(messages)

	if err := c.Close(); err != nil {
		output.Failf("Failed to close consumer: ", err)
	}
}
