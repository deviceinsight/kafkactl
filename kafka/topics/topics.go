package topics

import "github.com/Shopify/sarama"

type topic struct {
	Name       string      `json:"name"`
	Partitions []partition `json:"partitions,omitempty"`
}

type partition struct {
	Id           int32
	OldestOffset int64
	NewestOffset int64
	Leader       string  `json:",omitempty" yaml:",omitempty"`
	Replicas     []int32 `json:",omitempty" yaml:",omitempty"`
	ISRs         []int32 `json:",omitempty" yaml:",omitempty"`
}

func ReadTopic(client *sarama.Client, name string, readPartitions bool, readLeaders bool, readReplicas bool) (topic, error) {
	var (
		err error
		ps  []int32
		led *sarama.Broker
		top = topic{Name: name}
	)

	if !readPartitions {
		return top, nil
	}

	if ps, err = (*client).Partitions(name); err != nil {
		return top, err
	}

	for _, p := range ps {
		np := partition{Id: p}

		if np.OldestOffset, err = (*client).GetOffset(name, p, sarama.OffsetOldest); err != nil {
			return top, err
		}

		if np.NewestOffset, err = (*client).GetOffset(name, p, sarama.OffsetNewest); err != nil {
			return top, err
		}

		if readLeaders {
			if led, err = (*client).Leader(name, p); err != nil {
				return top, err
			}
			np.Leader = led.Addr()
		}

		if readReplicas {
			if np.Replicas, err = (*client).Replicas(name, p); err != nil {
				return top, err
			}

			if np.ISRs, err = (*client).InSyncReplicas(name, p); err != nil {
				return top, err
			}
		}

		top.Partitions = append(top.Partitions, np)
	}

	return top, nil
}
