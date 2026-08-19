package consume

import (
	"errors"
	"fmt"

	"github.com/IBM/sarama"
)

// DeserializationError marks a failure that originates from decoding a message
// (e.g. an invalid protobuf/avro payload or the absence of a suitable
// deserializer). It is distinguished from infrastructure errors so that
// consuming can optionally continue past individual undecodable messages.
type DeserializationError struct {
	err error
}

func (e *DeserializationError) Error() string {
	return e.err.Error()
}

func (e *DeserializationError) Unwrap() error {
	return e.err
}

// isDeserializationError reports whether err was caused by message
// deserialization rather than by the consumer infrastructure.
func isDeserializationError(err error) bool {
	var deserErr *DeserializationError
	return errors.As(err, &deserErr)
}

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
				return &DeserializationError{fmt.Errorf("failed to deserialize key: %w", err)}
			}
			break
		}

		if key == nil && flags.PrintKeys {
			return &DeserializationError{fmt.Errorf("can't find suitable deserializer for key")}
		}
	}

	// deserialize value
	for _, d := range *deserializer {
		if !d.CanDeserializeValue(consumerMsg, flags) {
			continue
		}

		value, err = d.DeserializeValue(consumerMsg)
		if err != nil {
			return &DeserializationError{fmt.Errorf("failed to deserialize value: %w", err)}
		}
		break
	}

	if value == nil {
		return &DeserializationError{fmt.Errorf("can't find suitable deserializer for value")}
	}

	// Apply filters - all filters must match (AND logic)
	if filter.IsActive() && !filter.Matches(consumerMsg, key, value) {
		return nil
	}

	// print message
	msg := newMessage(consumerMsg, flags, key, value)
	return printMessage(msg, flags)
}
