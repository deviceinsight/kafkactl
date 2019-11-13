package producer

import (
	"bufio"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/Shopify/sarama"
	"go.uber.org/ratelimit"
	"os"
	"os/signal"
	"strings"
)

type ProducerFlags struct {
	Partitioner				 string
	Partition					 int32
	Separator					 string
	Key								 string
	Value							 string
	KeySchemaVersion	 int
	ValueSchemaVersion int
	Silent						 bool
	RateInSeconds			 int
}

type ProducerOperation struct {
}

func (operation *ProducerOperation) Produce(topic string, flags ProducerFlags) {

	clientContext := operations.CreateClientContext()

	var (
		err				error
		client		sarama.Client
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

	producer, err := sarama.NewSyncProducer(clientContext.Brokers, config)
	if err != nil {
		output.Failf("Failed to open Kafka producer: %s", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			output.Warnf("Failed to close Kafka producer cleanly:", err)
		}
	}()

	var key string
	var value string

	if flags.Key != "" && flags.Separator != "" {
		output.Failf("parameters --key and --separator cannot be used together")
	} else if flags.Key != "" {
		key = flags.Key
	}

	if flags.Value != "" {
		// produce a single message
		message := serializer.Serialize([]byte(key), []byte(flags.Value), flags)
		partition, offset, err := producer.SendMessage(message)
		if err != nil {
			output.Failf("Failed to produce message: %s", err)
		} else if !flags.Silent {
			output.Infof("message produced (partition=%d\toffset=%d)\n", partition, offset)
		}
	} else if stdinAvailable() {

		cancel := make(chan struct{})

		go func() {
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Kill, os.Interrupt)
			<-signals
			close(cancel)
		}()

		messageCount := 0
		// print an empty line that will be replaced when updating the status
		output.Statusf("")

		var rl ratelimit.Limiter

		if flags.RateInSeconds == -1 {
			output.Debugf("No rate limiting was set, running without constraints")
			rl = ratelimit.NewUnlimited()
		} else {
			rl = ratelimit.New(flags.RateInSeconds) // per second
		}

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {

			select {
			case <-cancel:
				break
			default:
			}

			line, err := scanner.Text(), scanner.Err()
			if err != nil {
				output.Failf("Failed to read data from the standard input: %v", err)
			}

			if flags.Separator != "" {
				input := strings.SplitN(line, flags.Separator, 2)
				if len(input) < 2 {
					output.Failf("the provided input does not contain the separator %s", flags.Separator)
				}
				key = input[0]
				value = input[1]
			} else {
				value = line
			}

			messageCount += 1
			message := serializer.Serialize([]byte(key), []byte(value), flags)
			rl.Take()
			_, _, err = producer.SendMessage(message)
			if err != nil {
				output.Failf("Failed to produce message: %s", err)
			} else if !flags.Silent {
				if messageCount % 100 == 0 {
					output.Statusf("\r%d messages produced", messageCount)
				}
			}
		}

		output.Infof("\r%d messages produced", messageCount)
	} else {
		output.Failf("value is required, or you have to provide the value on stdin")
	}
}

func stdinAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
