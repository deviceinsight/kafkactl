package consume

import (
	"fmt"

	"github.com/IBM/sarama"
)

type MessageDeserializerChain []MessageDeserializer

func (deserializer *MessageDeserializerChain) Deserialize(consumerMsg *sarama.ConsumerMessage, flags Flags) error {

	var key, value *DeserializedData
	var err error

	// deserialize key
	if flags.PrintKeys {
		for _, d := range *deserializer {

			if !d.CanDeserializeKey(consumerMsg, flags) {
				continue
			}

			key, err = d.DeserializeKey(consumerMsg)
			if err != nil {
				return fmt.Errorf("failed to deserialize key: %w", err)
			}
			break
		}

		if key == nil {
			return fmt.Errorf("can't find suitable deserializer for key")
		}
	}

	// deserialize value
	for _, d := range *deserializer {
		if !d.CanDeserializeValue(consumerMsg, flags) {
			continue
		}

		value, err = d.DeserializeValue(consumerMsg)
		if err != nil {
			return fmt.Errorf("failed to deserialize value: %w", err)
		}
		break
	}

	if value == nil {
		return fmt.Errorf("can't find suitable deserializer for value")
	}

	// print message
	msg := newMessage(consumerMsg, flags, key, value)
	return printMessage(msg, flags)
}
