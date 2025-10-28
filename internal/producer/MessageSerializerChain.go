package producer

import (
	"github.com/IBM/sarama"
	"github.com/pkg/errors"

	"github.com/deviceinsight/kafkactl/v5/internal/producer/input"
)

type MessageSerializerChain struct {
	topic       string
	serializers []messageSerializer
}

func (serializer MessageSerializerChain) Serialize(msg input.Message, flags Flags) (*sarama.ProducerMessage, error) {
	recordHeaders, err := createRecordHeaders(flags)
	if err != nil {
		return nil, err
	}
	for k, v := range msg.Headers {
		recordHeaders = append(recordHeaders, sarama.RecordHeader{
			Key:   sarama.ByteEncoder(k),
			Value: sarama.ByteEncoder(v),
		})
	}
	message := &sarama.ProducerMessage{Topic: serializer.topic, Partition: flags.Partition, Headers: recordHeaders}

	if msg.Key != nil {
		bytes, err := serializer.serializeKey([]byte(*msg.Key), flags)
		if err != nil {
			return nil, err
		}
		message.Key = sarama.ByteEncoder(bytes)
	}
	var out []byte
	if msg.Value != nil {
		out = []byte(*msg.Value)
	}

	bytes, err := serializer.serializeValue(out, flags)
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
