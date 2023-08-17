package producer

import (
	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type MessageSerializerChain struct {
	topic       string
	serializers []MessageSerializer
}

func (serializer MessageSerializerChain) Serialize(key, value []byte, flags Flags) (*sarama.ProducerMessage, error) {
	for _, s := range serializer.serializers {
		canSerialize, err := s.CanSerialize(serializer.topic)
		if err != nil {
			return nil, err
		}

		if !canSerialize {
			continue
		}

		return s.Serialize(key, value, flags)
	}

	return nil, errors.Errorf("can't find suitable serializer")
}

func (serializer MessageSerializerChain) CanSerialize(topic string) (bool, error) {
	if serializer.topic != topic {
		return false, nil
	}

	for _, s := range serializer.serializers {
		canSerialize, err := s.CanSerialize(serializer.topic)
		if err != nil {
			return false, err
		}

		if canSerialize {
			return true, nil
		}
	}

	return false, nil
}
