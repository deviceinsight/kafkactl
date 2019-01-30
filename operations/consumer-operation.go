package operations

import (
	"encoding/base64"
	"encoding/hex"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/linkedin/goavro"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

type ConsumerFlags struct {
	PrintKeys       bool
	PrintTimestamps bool
	OutputFormat    string
	SchemaPath      string
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

type consumedMessage struct {
	Partition int32
	Offset    int64
	Key       *string `json:",omitempty" yaml:",omitempty"`
	Value     []byte
	Timestamp *time.Time `json:",omitempty" yaml:",omitempty"`
}

type ConsumerOperation struct {
}

func (operation *ConsumerOperation) Consume(topic string, flags ConsumerFlags) {

	clientContext := createClientContext()

	var (
		err       error
		client    sarama.Client
		topExists bool
	)

	if client, err = createClient(&clientContext); err != nil {
		output.Failf("failed to create client err=%v", err)
	}

	if topExists, err = topicExists(&client, topic); err != nil {
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

	c, err := sarama.NewConsumer(clientContext.brokers, nil)
	if err != nil {
		output.Failf("Failed to start consumer: %s", err)
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
			m := operation.newConsumedMessage(msg, flags)
			operation.printMessage(&m, flags)
		}
	}()

	wg.Wait()
	output.Infof("Done consuming topic: %s", topic)
	close(messages)

	if err := c.Close(); err != nil {
		output.Failf("Failed to close consumer: ", err)
	}
}

func (operation *ConsumerOperation) printMessage(msg *consumedMessage, flags ConsumerFlags) {
	if flags.OutputFormat == "" {
		var row []string

		if flags.PrintKeys {
			if msg.Key != nil {
				row = append(row, *msg.Key)
			} else {
				row = append(row, "")
			}
		}
		if flags.PrintTimestamps {
			if msg.Timestamp != nil {
				row = append(row, (*msg.Timestamp).Format(time.RFC3339))
			} else {
				row = append(row, "")
			}
		}

		var value string

		if flags.SchemaPath != "" {
			schema, err := ioutil.ReadFile(flags.SchemaPath)

			if err != nil {
				output.Failf("failed to read avro schema at '%s'", flags.SchemaPath)
			}

			codec, err := goavro.NewCodec(string(schema))

			if err != nil {
				output.Failf("failed to parse avro schema: %s", err)
			}

			native, _, err := codec.NativeFromBinary(msg.Value)
			if err != nil {
				output.Failf("failed to parse avro data: %s", err)
			}

			textual, err := codec.TextualFromNative(nil, native)
			if err != nil {
				output.Failf("failed to convert value to avro data: %s", err)
			}

			value = string(textual)
		} else {
			value = *encodeBytes(msg.Value, flags.EncodeValue)
		}

		row = append(row, string(value))

		output.PrintStrings(strings.Join(row[:], "#"))

	} else {
		output.PrintObject(msg, flags.OutputFormat)
	}
}

func (operation *ConsumerOperation) newConsumedMessage(m *sarama.ConsumerMessage, flags ConsumerFlags) consumedMessage {

	var key *string
	var timestamp *time.Time

	if flags.PrintKeys {
		key = encodeBytes(m.Key, flags.EncodeKey)
	}

	if flags.PrintTimestamps && !m.Timestamp.IsZero() {
		timestamp = &m.Timestamp
	}

	return consumedMessage{
		Partition: m.Partition,
		Offset:    m.Offset,
		Key:       key,
		Value:     m.Value,
		Timestamp: timestamp,
	}
}

func encodeBytes(data []byte, encoding string) *string {
	if data == nil {
		return nil
	}

	var str string
	switch encoding {
	case "hex":
		str = hex.EncodeToString(data)
	case "base64":
		str = base64.StdEncoding.EncodeToString(data)
	default:
		str = string(data)
	}

	return &str
}
