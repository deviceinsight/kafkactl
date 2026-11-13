package topic

import (
	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
)

func availableReplicaBrokers(client sarama.Client) []*sarama.Broker {
	var openedBrokers []*sarama.Broker
	for _, broker := range client.Brokers() {
		// Open first; Connected does not initiate connections.
		if err := broker.Open(client.Config()); err != nil && !errors.Is(err, sarama.ErrAlreadyConnected) {
			output.Warnf("unable to open broker %d (%s): %v", broker.ID(), broker.Addr(), err)
			continue
		}
		openedBrokers = append(openedBrokers, broker)
	}

	// Open dials asynchronously. Start all attempts before waiting for results.
	var availableBrokers []*sarama.Broker
	for _, broker := range openedBrokers {
		connected, err := broker.Connected()
		if err != nil {
			output.Warnf("unable to connect to broker %d (%s): %v", broker.ID(), broker.Addr(), err)
			continue
		}
		if connected {
			availableBrokers = append(availableBrokers, broker)
		} else {
			output.Warnf("broker %d (%s) is not available and will be excluded from replica assignment", broker.ID(), broker.Addr())
		}
	}
	return availableBrokers
}
