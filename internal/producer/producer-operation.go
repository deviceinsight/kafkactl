package producer

import (
	"bufio"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/internal"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"go.uber.org/ratelimit"
)

type Flags struct {
	Partitioner        string
	RequiredAcks       string
	MaxMessageBytes    int
	Partition          int32
	Separator          string
	LineSeparator      string
	File               string
	Key                string
	Value              string
	NullValue          bool
	Headers            []string
	KeySchemaVersion   int
	ValueSchemaVersion int
	KeyEncoding        string
	ValueEncoding      string
	Silent             bool
	RateInSeconds      int
	ProtoFiles         []string
	ProtoImportPaths   []string
	ProtosetFiles      []string
	KeyProtoType       string
	ValueProtoType     string
}

const DefaultMaxMessagesBytes = 1000000

type Operation struct {
}

func (operation *Operation) Produce(topic string, flags Flags) error {

	var (
		clientContext internal.ClientContext
		err           error
	)

	if clientContext, err = internal.CreateClientContext(); err != nil {
		return err
	}

	config, err := internal.CreateClientConfig(&clientContext)
	if err != nil {
		return err
	}

	// For implementation reasons, the SyncProducer requires `Producer.Return.Errors` and `Producer.Return.Successes` to
	// be set to true in its configuration.
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true

	if err = applyProducerConfigs(config, clientContext, flags); err != nil {
		return err
	}

	if flags.Separator != "" && (flags.Key != "" || flags.Value != "") {
		return errors.New("separator is used to split input from stdin/file. it cannot be used together with key or value")
	}

	serializers := MessageSerializerChain{topic: topic}

	if clientContext.AvroSchemaRegistry != "" {
		serializer, err := CreateAvroMessageSerializer(topic, clientContext.AvroSchemaRegistry, clientContext.AvroJSONCodec)
		if err != nil {
			return err
		}

		serializers.serializers = append(serializers.serializers, serializer)
	}

	if flags.KeyProtoType != "" || flags.ValueProtoType != "" {
		context := clientContext.Protobuf
		context.ProtosetFiles = append(flags.ProtosetFiles, context.ProtosetFiles...)
		context.ProtoFiles = append(flags.ProtoFiles, context.ProtoFiles...)
		context.ProtoImportPaths = append(flags.ProtoImportPaths, context.ProtoImportPaths...)

		serializer, err := CreateProtobufMessageSerializer(topic, context, flags.KeyProtoType, flags.ValueProtoType)
		if err != nil {
			return err
		}

		serializers.serializers = append(serializers.serializers, serializer)
	}

	serializers.serializers = append(serializers.serializers, DefaultMessageSerializer{topic: topic})

	output.Debugf("producer config: %+v", config.Producer)
	producer, err := sarama.NewSyncProducer(clientContext.Brokers, config)
	if err != nil {
		return errors.Wrap(err, "Failed to open Kafka producer")
	}
	defer func() {
		if err := producer.Close(); err != nil {
			output.Warnf("Failed to close Kafka producer cleanly:", err)
		}
	}()

	var key string
	var value string

	if flags.Key != "" && flags.Separator != "" {
		return errors.New("parameters --key and --separator cannot be used together")
	} else if flags.Key != "" {
		key = flags.Key
	}

	if flags.NullValue && flags.Value != "" {
		return errors.New("parameters --null-value and --value cannot be used together")
	}

	if flags.NullValue || flags.Value != "" {
		// produce a single message
		var message *sarama.ProducerMessage

		if flags.NullValue {
			message, err = serializers.Serialize([]byte(key), nil, flags)
		} else {
			message, err = serializers.Serialize([]byte(key), []byte(flags.Value), flags)
		}

		if err != nil {
			return errors.Wrap(err, "Failed to produce message")
		}
		partition, offset, err := producer.SendMessage(message)
		if err != nil {
			return errors.Wrap(err, "Failed to produce message")
		} else if !flags.Silent {
			output.Infof("message produced (partition=%d\toffset=%d)\n", partition, offset)
		}
	} else if flags.File != "" || stdinAvailable() {

		cancel := make(chan struct{})

		go func() {
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
			<-signals
			close(cancel)
		}()

		messageCount := 0
		keyColumnIdx := 0
		valueColumnIdx := 1
		columnCount := 2
		// print an empty line that will be replaced when updating the status
		output.Statusf("")

		var rl ratelimit.Limiter

		if flags.RateInSeconds == -1 {
			output.Debugf("No rate limiting was set, running without constraints")
			rl = ratelimit.NewUnlimited()
		} else {
			rl = ratelimit.New(flags.RateInSeconds) // per second
		}

		var inputReader io.Reader

		if flags.File != "" {
			inputReader, err = os.Open(flags.File)
			if err != nil {
				return errors.Errorf("unable to read input file %s: %v", flags.File, err)
			}
		} else {
			inputReader = os.Stdin
		}

		scanner := bufio.NewScanner(inputReader)
		scanner.Buffer(make([]byte, 0, config.Producer.MaxMessageBytes), config.Producer.MaxMessageBytes)

		if len(flags.LineSeparator) > 0 && flags.LineSeparator != "\n" {
			scanner.Split(splitAt(convertControlChars(flags.LineSeparator)))
		}

		for scanner.Scan() {

			select {
			case <-cancel:
				break
			default:
			}

			line, err := scanner.Text(), scanner.Err()
			if err != nil {
				return failWithMessageCount(messageCount, "Failed to read data from the standard input: %v", err)
			}

			if strings.TrimSpace(line) == "" {
				continue
			}

			if flags.Separator != "" {
				input := strings.Split(line, convertControlChars(flags.Separator))
				if len(input) < 2 {
					return failWithMessageCount(messageCount, "the provided input does not contain the separator %s", flags.Separator)
				} else if len(input) == 3 && messageCount == 0 {
					keyColumnIdx, valueColumnIdx, columnCount, err = resolveColumns(input)
					if err != nil {
						return failWithMessageCount(messageCount, err.Error())
					}
				} else if len(input) != columnCount {
					return failWithMessageCount(messageCount, "line contains unexpected amount of separators:\n%s", line)
				}
				key = input[keyColumnIdx]
				value = input[valueColumnIdx]
			} else {
				value = line
			}

			messageCount++
			message, err := serializers.Serialize([]byte(key), []byte(value), flags)
			if err != nil {
				return errors.Wrap(err, "Failed to produce message")
			}
			rl.Take()
			_, _, err = producer.SendMessage(message)
			if err != nil {
				return failWithMessageCount(messageCount, "Failed to produce message: %s", err)
			} else if !flags.Silent {
				if messageCount%100 == 0 {
					output.Statusf("\r%d messages produced", messageCount)
				}
			}
		}

		if scanner.Err() != nil {
			return errors.Wrap(scanner.Err(), "error reading input (try specifying --max-message-bytes when producing long messages)")
		}

		output.Infof("\r%d messages produced", messageCount)
	} else {
		return errors.New("value is required, or you have to provide the value on stdin")
	}
	return nil
}

func applyProducerConfigs(config *sarama.Config, clientContext internal.ClientContext, flags Flags) error {

	var err error

	partitioner := clientContext.Producer.Partitioner
	if flags.Partitioner != "" {
		partitioner = flags.Partitioner
	}

	if config.Producer.Partitioner, err = parsePartitioner(partitioner, flags); err != nil {
		return err
	}

	requiredAcks := clientContext.Producer.RequiredAcks
	if flags.RequiredAcks != "" {
		requiredAcks = flags.RequiredAcks
	}

	if config.Producer.RequiredAcks, err = parseRequiredAcks(requiredAcks); err != nil {
		return err
	}

	maxMessageBytes := DefaultMaxMessagesBytes
	if flags.MaxMessageBytes > 0 {
		maxMessageBytes = flags.MaxMessageBytes
	}

	config.Producer.MaxMessageBytes = maxMessageBytes

	return nil
}

func parseRequiredAcks(requiredAcks string) (sarama.RequiredAcks, error) {
	switch requiredAcks {
	case "NoResponse":
		return sarama.NoResponse, nil
	case "WaitForAll":
		return sarama.WaitForAll, nil
	case "WaitForLocal":
		fallthrough
	case "":
		return sarama.WaitForLocal, nil
	default:
		return sarama.WaitForLocal, errors.Errorf("unknown required-acks setting: %s", requiredAcks)
	}
}

func parsePartitioner(partitioner string, flags Flags) (sarama.PartitionerConstructor, error) {
	switch partitioner {
	case "":
		if flags.Partition >= 0 {
			return sarama.NewManualPartitioner, nil
		}
		return NewJVMCompatiblePartitioner, nil
	case "murmur2":
		// https://github.com/IBM/sarama/issues/1424
		return NewJVMCompatiblePartitioner, nil
	case "hash":
		return sarama.NewHashPartitioner, nil
	case "hash-ref":
		return sarama.NewReferenceHashPartitioner, nil
	case "random":
		return sarama.NewRandomPartitioner, nil
	case "manual":
		if flags.Partition == -1 {
			return nil, errors.New("partition is required when partitioning manually")
		}
		return sarama.NewManualPartitioner, nil
	default:
		return nil, errors.Errorf("partitioner %s not supported", flags.Partitioner)
	}
}

func convertControlChars(value string) string {
	value = strings.Replace(value, "\\n", "\n", -1)
	value = strings.Replace(value, "\\r", "\r", -1)
	value = strings.Replace(value, "\\t", "\t", -1)
	return value
}

func resolveColumns(line []string) (keyColumnIdx, valueColumnIdx, columnCount int, err error) {
	if isTimestamp(line[0]) {
		output.Warnf("assuming column 0 to be message timestamp. Column will be ignored")
		return 1, 2, 3, nil
	} else if isTimestamp(line[1]) {
		output.Warnf("assuming column 1 to be message timestamp. Column will be ignored")
		return 0, 2, 3, nil
	} else {
		return -1, -1, -1, errors.Errorf("line contains unexpected amount of separators:\n%s", line)
	}
}

func isTimestamp(value string) bool {
	_, e := time.Parse(time.RFC3339, value)
	return e == nil
}

func failWithMessageCount(messageCount int, errorMessage string, args ...interface{}) error {
	output.Infof("\r%d messages produced", messageCount)
	return errors.Errorf(errorMessage, args...)
}

func stdinAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func splitAt(delimiter string) func(data []byte, atEOF bool) (advance int, token []byte, err error) {

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {

		// Return nothing if at end of file and no data passed
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// Find the index of the input of the separator delimiter
		if i := strings.Index(string(data), delimiter); i >= 0 {
			return i + len(delimiter), data[0:i], nil
		}

		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}

		// Request more data.
		return 0, nil, nil
	}
}
