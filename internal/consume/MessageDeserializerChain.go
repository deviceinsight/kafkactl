package consume

import (
	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type MessageDeserializerChain []MessageDeserializer

func (deserializer MessageDeserializerChain) Deserialize(msg *sarama.ConsumerMessage, flags Flags) error {
	for _, d := range deserializer {
		canDeserialize, err := d.CanDeserialize(msg.Topic)
		if err != nil {
			return err
		}

		if !canDeserialize {
			continue
		}

		return d.Deserialize(msg, flags)
	}

	return errors.Errorf("can't find suitable deserializer")
}

func (deserializer MessageDeserializerChain) CanDeserialize(topic string) (bool, error) {
	for _, d := range deserializer {
		canDeserialize, err := d.CanDeserialize(topic)
		if err != nil {
			return false, err
		}

		if canDeserialize {
			return true, nil
		}
	}

	return false, nil
}
