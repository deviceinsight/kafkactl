package consume

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/output"
	"golang.org/x/sync/errgroup"
)

type GroupConsumer struct {
	topic               string
	group               string
	client              *sarama.Client
	consumerGroupClient *sarama.ConsumerGroup
	errorGroup          *errgroup.Group
}

func CreateGroupConsumer(client *sarama.Client, topic string, group string) (*GroupConsumer, error) {

	return &GroupConsumer{
		topic:  topic,
		group:  group,
		client: client,
	}, nil
}

func (c *GroupConsumer) Start(ctx context.Context, flags Flags, messages chan<- *sarama.ConsumerMessage) error {

	if flags.FromBeginning {
		(*c.client).Config().Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	consumerGroupClient, err := sarama.NewConsumerGroupFromClient(c.group, *c.client)
	if err != nil {
		return err
	}
	c.consumerGroupClient = &consumerGroupClient

	var groupHandler = groupHandler{
		messages: messages,
		ready:    make(chan bool),
	}
	c.errorGroup, ctx = errgroup.WithContext(ctx)

	c.errorGroup.Go(func() error {
		for {
			if err := consumerGroupClient.Consume(ctx, []string{c.topic}, &groupHandler); err != nil {
				return err
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return nil
			}
			groupHandler.ready = make(chan bool)
		}
	})

	<-groupHandler.ready
	output.Debugf("group consumer initialized")
	return nil
}

func (c *GroupConsumer) Wait() error {
	output.Debugf("waiting for group consumer")
	return c.errorGroup.Wait()
}

func (c *GroupConsumer) Close() error {
	output.Debugf("closing consumer")
	return (*c.consumerGroupClient).Close()
}

type groupHandler struct {
	messages chan<- *sarama.ConsumerMessage
	ready    chan bool
}

func (handler *groupHandler) Setup(sarama.ConsumerGroupSession) error {
	close(handler.ready)
	return nil
}

func (handler *groupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (handler *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for message := range claim.Messages() {
		if message != nil {
			handler.messages <- message
		}
		session.MarkMessage(message, "")
	}

	return nil
}
