package consumergroupoffsets

import "github.com/IBM/sarama"

type OffsetSettingConsumer struct {
	Topic            string
	PartitionOffsets map[int32]PartitionOffset

	ready chan struct{}
}

func (s *OffsetSettingConsumer) Setup(session sarama.ConsumerGroupSession) error {
	s.ready = make(chan struct{})

	for partition, offset := range s.PartitionOffsets {
		session.MarkOffset(s.Topic, partition, offset.Offset, offset.Metadata)
	}

	return nil
}

func (s *OffsetSettingConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	close(s.ready)
	return nil
}

func (s *OffsetSettingConsumer) ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error {
	return nil
}
