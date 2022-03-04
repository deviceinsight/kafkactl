package consume

import (
	"context"

	"github.com/Shopify/sarama"
)

type Consumer interface {
	Start(ctx context.Context, flags Flags, messages chan<- *sarama.ConsumerMessage, stopConsumers <-chan bool) error
	Wait() error
	Close() error
}
