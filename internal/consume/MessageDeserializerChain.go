package consume

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/IBM/sarama"
)

type MessageDeserializerChain []MessageDeserializer

func (deserializer *MessageDeserializerChain) Deserialize(consumerMsg *sarama.ConsumerMessage, flags Flags) error {

	var key, value *DeserializedData
	var err error

	// determine whether we need to deserialize the key: either because keys are requested to be printed
	// or because a filter keyword/regex should be applied to keys
	needKey := flags.PrintKeys || flags.FilterKeyword != ""

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

	filterMode := FilterMode(flags.FilterMode)
	if flags.FilterMode == "" {
		filterMode = FilterModeAny
	}
	allowFilteringMessage := filterMode == FilterModeAny || filterMode == FilterModeMessage
	allowFilteringKey := filterMode == FilterModeAny || filterMode == FilterModeKey

	// If a filter keyword is present, attempt to match key and/or value now and only print
	// when a match occurs. Only consider matching if the deserialized bytes are textual (utf8.Valid).
	if flags.FilterKeyword != "" {
		matched := false

		// check value first
		if allowFilteringMessage && value != nil && value.data != nil && utf8.Valid(value.data) {
			vs := string(value.data)
			if flags.FilterByRegex {
				if flags.filterRegexp != nil && flags.filterRegexp.MatchString(vs) {
					matched = true
				}
			} else {
				if strings.Contains(vs, flags.FilterKeyword) {
					matched = true
				}
			}
		}

		// check key if not matched yet
		if !matched && allowFilteringKey && key != nil && key.data != nil && utf8.Valid(key.data) {
			ks := string(key.data)
			if flags.FilterByRegex {
				if flags.filterRegexp != nil && flags.filterRegexp.MatchString(ks) {
					matched = true
				}
			} else {
				if strings.Contains(ks, flags.FilterKeyword) {
					matched = true
				}
			}
		}

		if !matched {
			// skip printing non-matching message
			return nil
		}
	}

	// print message
	msg := newMessage(consumerMsg, flags, key, value)
	return printMessage(msg, flags)
}
