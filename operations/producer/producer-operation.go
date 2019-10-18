package producer

import (
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/operations"
	"github.com/deviceinsight/kafkactl/output"
	"bufio"
//"io/ioutil"
	"os"
	"strings"
//"time"
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
	RateInSeconds      int
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

  if flags.RateInSeconds == -1 {
			output.Infof("No rate limiting was set, running without constraints")
  } else {
// Basic rate limiting from google
// taken from https://github.com/golang/go/wiki/RateLimiting
//
//    rate := time.Second / flags.RateInSeconds
//    burstLimit := 100
//    tick := time.NewTicker(rate)
//    defer tick.Stop()
//    throttle := make(chan time.Time, burstLimit)
//    go func() {
//      for t := range tick.C {
//        select {
//          case throttle <- t:
//          default:
//        }
//      }  // does not exit after tick.Stop()
//    }()
//    for req := range requests {
//      <-throttle  // rate limit our Service.Method RPCs
//      go client.Call("Service.Method", req, ...)
//    }
  }

	var key []byte
	var value []byte

	if flags.Key != "" {
		key = []byte(flags.Key)
	}

	if flags.Value != "" {
		value = []byte(flags.Value)
	} else if stdinAvailable() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			bytes, err := scanner.Text(), scanner.Err()
			if err != nil {
				output.Failf("Failed to read data from the standard input: %s", err)
			}

			if flags.Separator != "" {
				// In this case we need to buffer the requests or use another seperator to seperate messages
				input := strings.SplitN(string(bytes[:]), flags.Separator, 2)
				if len(input) < 2 {
					output.Failf("the provided input does not contain the separator %s", flags.Separator)
			}
				key = []byte(input[0])
				value = []byte(input[1])
			} else {
        value = []byte(bytes)
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

		if scanner.Err() != nil {
			output.Failf("Failed to read data from the standard input: %s", err)
		}
	} else {
		output.Failf("value is required, or you have to provide the value on stdin")
	}

}

func stdinAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
