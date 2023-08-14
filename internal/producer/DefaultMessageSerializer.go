package producer

import (
	"github.com/IBM/sarama"
)

type DefaultMessageSerializer struct {
	topic string
}

func (serializer DefaultMessageSerializer) CanSerialize(_ string) (bool, error) {
	return true, nil
}

func (serializer DefaultMessageSerializer) Serialize(key, value []byte, flags Flags) (*sarama.ProducerMessage, error) {

	recordHeaders, err := createRecordHeaders(flags)
	if err != nil {
		return nil, err
	}
	message := &sarama.ProducerMessage{Topic: serializer.topic, Partition: flags.Partition, Headers: recordHeaders}

	if key != nil {
		bytes, err := decodeBytes(key, flags.KeyEncoding)
		if err != nil {
			return nil, err
		}
		message.Key = sarama.ByteEncoder(bytes)
	}

	bytes, err := decodeBytes(value, flags.ValueEncoding)
	if err != nil {
		return nil, err
	}
	message.Value = sarama.ByteEncoder(bytes)

	return message, nil
}
