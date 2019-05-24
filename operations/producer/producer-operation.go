package producer

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"io/ioutil"
	"os"
	"strings"
)

type ProducerFlags struct {
	Partitioner        string
	Partition          int32
	Separator          string
	Key                string
	Value              string
	KeySchemaVersion   int
	ValueSchemaVersion int
	Silent             bool
}

type ProducerOperation struct {
}

func (operation *ProducerOperation) Produce(topic string, flags ProducerFlags) {

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

	config := operations.CreateClientConfig(&clientContext)
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

	if flags.Separator != "" && (flags.Key != "" || flags.Value != "") {
		output.Failf("separator is used to split stdin. it cannot be used together with key or value")
	}

	var serializer MessageSerializer

	if clientContext.AvroSchemaRegistry != "" {
		serializer = CreateAvroMessageSerializer(topic, clientContext.AvroSchemaRegistry)
	} else {
		serializer = DefaultMessageSerializer{topic: topic}
	}

	var key []byte
	var value []byte

	if flags.Key != "" {
		key = []byte(flags.Key)
	}

	if flags.Value != "" {
		value = []byte(flags.Value)
	} else if stdinAvailable() {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			output.Failf("Failed to read data from the standard input: %s", err)
		}

		if len(bytes) > 0 && bytes[len(bytes)-1] == 10 {
			// remove trailing return
			bytes = bytes[:len(bytes)-1]
		}

		if flags.Separator != "" {
			input := strings.SplitN(string(bytes[:]), flags.Separator, 2)
			if len(input) < 2 {
				output.Failf("the provided input does not contain the separator %s", flags.Separator)
			}
			key = []byte(input[0])
			value = []byte(input[1])
		} else {
			value = bytes
		}
	} else {
		output.Failf("value is required, or you have to provide the value on stdin")
	}

	message := serializer.Serialize(key, value, flags)

	producer, err := sarama.NewSyncProducer(clientContext.Brokers, config)
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
		output.Infof("topic=%s\tpartition=%d\toffset=%d\n", topic, partition, offset)
	}
}

func stdinAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
