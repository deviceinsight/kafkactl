package consume

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/deviceinsight/kafkactl/v5/internal/helpers/protobuf"

	"github.com/deviceinsight/kafkactl/v5/internal/helpers"

	"golang.org/x/sync/errgroup"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
)

type Flags struct {
	PrintPartitions     bool
	PrintKeys           bool
	PrintTimestamps     bool
	PrintSchema         bool
	PrintHeaders        bool
	PrintAll            bool
	OutputFormat        string
	Separator           string
	Group               string
	Partitions          []int
	Offsets             []string
	FromBeginning       bool
	FromTimestamp       string
	ToTimestamp         string
	Tail                int
	Exit                bool
	MaxMessages         int64
	EncodeValue         string
	EncodeKey           string
	ProtoFiles          []string
	ProtoImportPaths    []string
	ProtosetFiles       []string
	ProtoMarshalOptions []string
	KeyProtoType        string
	ValueProtoType      string
	IsolationLevel      string

	FilterKey    string
	FilterValue  string
	FilterHeader map[string]string
}

type ConsumedMessage struct {
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Timestamp *time.Time
}

type Operation struct{}

func (operation *Operation) Consume(topic string, flags Flags) error {
	var (
		clientContext internal.ClientContext
		err           error
		client        sarama.Client
		topExists     bool
	)

	if clientContext, err = internal.CreateClientContext(); err != nil {
		return err
	}

	config, err := internal.CreateClientConfig(&clientContext)
	if err != nil {
		return err
	}

	if err = applyConsumerConfigs(config, clientContext, flags); err != nil {
		return err
	}

	if client, err = sarama.NewClient(clientContext.Brokers, config); err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	if topExists, err = internal.TopicExists(&client, topic); err != nil {
		return errors.Wrap(err, "failed to read topics")
	}

	if !topExists {
		return errors.Errorf("topic '%s' does not exist", topic)
	}

	var schemaRegistryClient *internal.CachingSchemaRegistry

	if clientContext.SchemaRegistry.URL != "" {
		schemaRegistryClient, err = internal.CreateCachingSchemaRegistry(&clientContext)
		if err != nil {
			return err
		}
	}

	var deserializers MessageDeserializerChain
	var protobufConfig internal.ProtobufConfig

	if protobufConfig, err = addFlagsToProtobufConfig(clientContext.Protobuf, flags); err != nil {
		return err
	}

	if schemaRegistryClient != nil {
		avroDeserializer := AvroMessageDeserializer{topic: topic, registry: schemaRegistryClient, jsonCodec: clientContext.Avro.JSONCodec}
		protobufDeserializer := RegistryProtobufMessageDeserializer{config: protobufConfig, registry: schemaRegistryClient}
		deserializers = append(deserializers, &avroDeserializer, &protobufDeserializer)
	}

	deserializer, err := CreateProtobufMessageDeserializer(protobufConfig, protoreflect.FullName(flags.KeyProtoType), protoreflect.FullName(flags.ValueProtoType))
	if err != nil {
		return err
	}

	deserializers = append(deserializers, deserializer)
	deserializers = append(deserializers, &DefaultMessageDeserializer{})

	if flags.Group != "" {
		if flags.Exit {
			return errors.New("parameters --group and --exit cannot be used together")
		}

		if flags.Tail > 0 {
			return errors.New("parameters --group and --tail cannot be used together")
		}

		if len(flags.Partitions) > 0 {
			return errors.New("parameters --group and --partitions cannot be used together")
		}

		if len(flags.Offsets) > 0 {
			return errors.New("parameters --group and --offset cannot be used together")
		}
	}

	messages := make(chan *sarama.ConsumerMessage)
	stopConsumers := make(chan bool)

	var consumer Consumer

	if flags.Group == "" {
		consumer, err = CreatePartitionConsumer(&client, topic, flags.Partitions)
	} else {
		consumer, err = CreateGroupConsumer(&client, topic, flags.Group)
	}

	if err != nil {
		return err
	}

	output.Debugf("Start consuming topic: %s", topic)

	ctx := helpers.CreateTerminalContext()

	if err := consumer.Start(ctx, flags, messages, stopConsumers); err != nil {
		return errors.Wrap(err, "Failed to start consumer")
	}

	messageFilter, err := NewMessageFilter(flags.FilterKey, flags.FilterValue, flags.FilterHeader)
	if err != nil {
		return err
	}

	deserializationGroup := deserializeMessages(ctx, flags, messages, stopConsumers, deserializers, messageFilter)

	if err := consumer.Wait(); err != nil {
		return errors.Wrap(err, "Failed while waiting for consumer")
	}

	close(messages)

	output.Debugf("waiting for deserialization")
	if err := deserializationGroup.Wait(); err != nil {
		return errors.Wrap(err, "Error during deserialization")
	}
	output.Debugf("deserialization finished")

	if err := consumer.Close(); err != nil {
		return errors.Wrap(err, "Failed to close consumer")
	}

	return nil
}

func addFlagsToProtobufConfig(protobufConfig internal.ProtobufConfig, flags Flags) (internal.ProtobufConfig, error) {

	protobufConfig.ProtosetFiles = append(flags.ProtosetFiles, protobufConfig.ProtosetFiles...)
	protobufConfig.ProtoFiles = append(flags.ProtoFiles, protobufConfig.ProtoFiles...)
	protobufConfig.ProtoImportPaths = append(flags.ProtoImportPaths, protobufConfig.ProtoImportPaths...)

	var err error

	protobufConfig.MarshalOptions, err = protobuf.ParseMarshalOptions(flags.ProtoMarshalOptions, protobufConfig.MarshalOptions)
	if err != nil {
		return protobufConfig, fmt.Errorf("error parsing protobuf marshal options: %w", err)
	}

	return protobufConfig, nil
}

func applyConsumerConfigs(config *sarama.Config, clientContext internal.ClientContext, flags Flags) error {
	var err error

	isolationLevel := clientContext.Consumer.IsolationLevel
	if flags.IsolationLevel != "" {
		isolationLevel = flags.IsolationLevel
	}

	if config.Consumer.IsolationLevel, err = parseIsolationLevel(isolationLevel); err != nil {
		return err
	}

	output.Debugf("using isolationLevel=%v", config.Consumer.IsolationLevel)

	return nil
}

func parseIsolationLevel(isolationLevel string) (sarama.IsolationLevel, error) {
	switch strings.ToLower(isolationLevel) {
	case "":
		return sarama.ReadCommitted, nil
	case "readcommitted":
		return sarama.ReadCommitted, nil
	case "readuncommitted":
		return sarama.ReadUncommitted, nil
	default:
		return sarama.ReadCommitted, errors.Errorf("isolationLevel=%s not supported", isolationLevel)
	}
}

func deserializeMessages(ctx context.Context, flags Flags, messages <-chan *sarama.ConsumerMessage,
	stopConsumers chan<- bool, deserializers MessageDeserializerChain, filter *MessageFilter,
) *errgroup.Group {
	errorGroup, _ := errgroup.WithContext(ctx)

	if flags.Tail > 0 {
		errorGroup.Go(func() error {
			sortedMessages := make([]*sarama.ConsumerMessage, 0)

			for msg := range messages {
				sortedMessages = insertSorted(sortedMessages, msg)
				if len(sortedMessages) > flags.Tail {
					sortedMessages = sortedMessages[:flags.Tail]
				}
			}
			lastIndex := len(sortedMessages) - 1
			for i := range sortedMessages {
				err := deserializers.Deserialize(sortedMessages[lastIndex-i], flags, filter)
				if err != nil {
					return err
				}
			}

			return nil
		})
	} else {
		// just print the messages
		errorGroup.Go(func() error {
			var messageCount int64
			var err error

			for msg := range messages {
				err = deserializers.Deserialize(msg, flags, filter)
				messageCount++
				if err != nil {
					close(stopConsumers)
					break
				}
				if flags.MaxMessages > 0 && messageCount >= flags.MaxMessages {
					close(stopConsumers)
					break
				}
			}

			// drop remaining messages after break
			for range messages {
				output.Debugf("drop message")
			}

			return err
		})
	}

	return errorGroup
}

func insertSorted(messages []*sarama.ConsumerMessage, message *sarama.ConsumerMessage) []*sarama.ConsumerMessage {
	index := sort.Search(len(messages), func(i int) bool {
		return messages[i].Timestamp.Before(message.Timestamp)
	})
	messages = append(messages, nil)
	copy(messages[index+1:], messages[index:])
	messages[index] = message
	return messages
}
