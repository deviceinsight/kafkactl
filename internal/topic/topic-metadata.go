package topic

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

func waitForTopicPartitions(client sarama.Client, name string, count int32) error {
	deadline := time.Now().Add(client.Config().Admin.Timeout)
	for {
		if err := client.RefreshMetadata(name); err != nil {
			return err
		}
		partitions, err := client.Partitions(name)
		if err != nil {
			return err
		}
		if len(partitions) >= int(count) {
			return nil
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return errors.Errorf("topic %q has %d partitions in metadata, expected at least %d", name, len(partitions), count)
		}
		time.Sleep(min(50*time.Millisecond, remaining))
	}
}
