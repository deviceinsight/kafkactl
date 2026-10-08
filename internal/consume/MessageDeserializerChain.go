package consume

import (
	"fmt"

	"github.com/IBM/sarama"
)

type MessageDeserializerChain []MessageDeserializer

func (deserializer *MessageDeserializerChain) Deserialize(consumerMsg *sarama.ConsumerMessage, flags Flags, filter *MessageFilter) error {

	var key, value *DeserializedData
	var err error

	// determine whether we need to deserialize the key: either because keys are requested to be printed
	// or because a filter is applied to keys
	needKey := flags.PrintKeys || flags.FilterKey != ""

	// deserialize key
	if needKey {
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

		if key == nil && flags.PrintKeys {
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

	// Apply filters - all filters must match (AND logic)
	if filter.IsActive() && !filter.Matches(consumerMsg, key, value) {
		return nil
	}

	// print message
	msg := newMessage(consumerMsg, flags, key, value)
	return printMessage(msg, flags)
}
