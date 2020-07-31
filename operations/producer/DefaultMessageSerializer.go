package producer

import (
	"github.com/Shopify/sarama"
)

type DefaultMessageSerializer struct {
	topic string
}

func (serializer DefaultMessageSerializer) CanSerialize(_ string) (bool, error) {
	return true, nil
}

func (serializer DefaultMessageSerializer) Serialize(key, value []byte, flags ProducerFlags) (*sarama.ProducerMessage, error) {

	recordHeaders, err := createRecordHeaders(flags)
	if err != nil {
		return nil, err
	}
	message := &sarama.ProducerMessage{Topic: serializer.topic, Partition: flags.Partition, Headers: recordHeaders}

	if key != nil {
		message.Key = sarama.ByteEncoder(decodeBytes(key, flags.KeyEncoding))
	}

	message.Value = sarama.ByteEncoder(decodeBytes(value, flags.ValueEncoding))

	return message, nil
}
