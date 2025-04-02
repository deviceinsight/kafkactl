package producer

import (
	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type MessageSerializerChain struct {
	topic       string
	serializers []messageSerializer
}

func (serializer MessageSerializerChain) Serialize(key, value []byte, flags Flags) (*sarama.ProducerMessage, error) {
	recordHeaders, err := createRecordHeaders(flags)
	if err != nil {
		return nil, err
	}
	message := &sarama.ProducerMessage{Topic: serializer.topic, Partition: flags.Partition, Headers: recordHeaders}

	if key != nil {
		bytes, err := serializer.serializeKey(key, flags)
		if err != nil {
			return nil, err
		}
		message.Key = sarama.ByteEncoder(bytes)
	}

	bytes, err := serializer.serializeValue(value, flags)
	if err != nil {
		return nil, err
	}
	message.Value = sarama.ByteEncoder(bytes)

	return message, nil
}

func (serializer MessageSerializerChain) serializeValue(value []byte, flags Flags) ([]byte, error) {
	for _, s := range serializer.serializers {
		canSerialize, err := s.CanSerializeValue(serializer.topic)
		if err != nil {
			return nil, err
		}

		if !canSerialize {
			continue
		}
		decodedValue, err := decodeBytes(value, flags.ValueEncoding)
		if err != nil {
			return nil, err
		}

		return s.SerializeValue(decodedValue, flags)
	}

	return nil, errors.Errorf("can't find suitable serializer")
}

func (serializer MessageSerializerChain) serializeKey(key []byte, flags Flags) ([]byte, error) {
	for _, s := range serializer.serializers {
		canSerialize, err := s.CanSerializeKey(serializer.topic)
		if err != nil {
			return nil, err
		}

		if !canSerialize {
			continue
		}
		decodedKey, err := decodeBytes(key, flags.KeyEncoding)
		if err != nil {
			return nil, err
		}

		return s.SerializeKey(decodedKey, flags)
	}

	return nil, errors.Errorf("can't find suitable serializer")
}
