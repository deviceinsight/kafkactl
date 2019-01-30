package operations

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/linkedin/goavro"
	"io/ioutil"
	"os"
	"strings"
)

type ProducerFlags struct {
	Partitioner string
	Partition   int32
	Separator   string
	Key         string
	Value       string
	SchemaPath  string
	Silent      bool
}

type ProducerOperation struct {
}

func (operation *ProducerOperation) Produce(topic string, flags ProducerFlags) {

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

	config := createClientConfig(&clientContext)
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	switch flags.Partitioner {
	case "":
		if flags.Partition >= 0 {
			config.Producer.Partitioner = sarama.NewManualPartitioner
		} else {
			config.Producer.Partitioner = sarama.NewHashPartitioner
		}
	case "hash":
		config.Producer.Partitioner = sarama.NewHashPartitioner
	case "random":
		config.Producer.Partitioner = sarama.NewRandomPartitioner
	case "manual":
		config.Producer.Partitioner = sarama.NewManualPartitioner
		if flags.Partition == -1 {
			output.Failf("partition is required when partitioning manually")
		}
	default:
		output.Failf("Partitioner %s not supported.", flags.Partitioner)
	}

	message := &sarama.ProducerMessage{Topic: topic, Partition: flags.Partition}

	if flags.Separator != "" && (flags.Key != "" || flags.Value != "") {
		output.Failf("separator is used to split stdin. it cannot be used together with key or value")
	}

	if flags.Key != "" {
		message.Key = sarama.StringEncoder(flags.Key)
	}

	var value []byte

	if flags.Value != "" {
		value = []byte(flags.Value)
	} else if stdinAvailable() {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			output.Failf("Failed to read data from the standard input: %s", err)
		}
		if flags.Separator != "" {
			input := strings.SplitN(string(bytes[:]), flags.Separator, 2)
			if len(input) < 2 {
				output.Failf("the provided input does not contain the separator %s", flags.Separator)
			}
			message.Key = sarama.StringEncoder(input[0])
			value = []byte(input[1])
		} else {
			value = bytes
		}
	} else {
		output.Failf("value is required, or you have to provide the value on stdin")
	}

	if flags.SchemaPath != "" {
		schema, err := ioutil.ReadFile(flags.SchemaPath)

		if err != nil {
			output.Failf("failed to read avro schema at '%s'", flags.SchemaPath)
		}

		codec, err := goavro.NewCodec(string(schema))

		if err != nil {
			output.Failf("failed to parse avro schema: %s", err)
		}

		native, _, err := codec.NativeFromTextual(value)
		if err != nil {
			output.Failf("failed to convert value to avro data: %s", err)
		}

		binary, err := codec.BinaryFromNative(nil, native)
		if err != nil {
			output.Failf("failed to convert value to avro data: %s", err)
		}

		message.Value = sarama.ByteEncoder(binary)
	} else {
		message.Value = sarama.ByteEncoder(value)
	}

	producer, err := sarama.NewSyncProducer(clientContext.brokers, config)
	if err != nil {
		output.Failf("Failed to open Kafka producer: %s", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			output.Warnf("Failed to close Kafka producer cleanly:", err)
		}
	}()

	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		output.Failf("Failed to produce message: %s", err)
	} else if !flags.Silent {
		fmt.Printf("topic=%s\tpartition=%d\toffset=%d\n", topic, partition, offset)
	}
}

func stdinAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
