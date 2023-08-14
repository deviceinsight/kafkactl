package consume

import (
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/internal/helpers/protobuf"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

type ProtobufMessageDeserializer struct {
	keyDescriptor   *desc.MessageDescriptor
	valueDescriptor *desc.MessageDescriptor
}

func CreateProtobufMessageDeserializer(context protobuf.SearchContext, keyType, valueType string) (*ProtobufMessageDeserializer, error) {
	valueDescriptor := protobuf.ResolveMessageType(context, valueType)
	if valueDescriptor == nil {
		return nil, errors.Errorf("value message type %q not found in provided files", valueType)
	}

	keyDescriptor := protobuf.ResolveMessageType(context, keyType)
	if valueDescriptor == nil && keyType != "" {
		return nil, errors.Errorf("key message type %q not found in provided files", valueType)
	}

	return &ProtobufMessageDeserializer{
		keyDescriptor:   keyDescriptor,
		valueDescriptor: valueDescriptor,
	}, nil
}

type protobufMessage struct {
	Partition int32
	Offset    int64
	Headers   map[string]string `json:",omitempty" yaml:",omitempty"`
	Key       *string           `json:",omitempty" yaml:",omitempty"`
	Value     *string
	Timestamp *time.Time `json:",omitempty" yaml:",omitempty"`
}

func (deserializer ProtobufMessageDeserializer) newProtobufMessage(msg *sarama.ConsumerMessage, flags Flags) (protobufMessage, error) {
	var err error

	ret := protobufMessage{
		Partition: msg.Partition,
		Offset:    msg.Offset,
	}

	if flags.PrintTimestamps && !msg.Timestamp.IsZero() {
		ret.Timestamp = &msg.Timestamp
	}

	if flags.PrintHeaders {
		ret.Headers = encodeRecordHeaders(msg.Headers)
	}

	ret.Value, err = decodeProtobuf(msg.Value, deserializer.valueDescriptor, flags.EncodeValue)
	if err != nil {
		return protobufMessage{}, errors.Wrap(err, "value decode failed")
	}

	if flags.PrintKeys {
		if deserializer.keyDescriptor != nil {
			ret.Key, err = decodeProtobuf(msg.Key, deserializer.keyDescriptor, flags.EncodeKey)
			if err != nil {
				return protobufMessage{}, errors.Wrap(err, "key decode failed")
			}
		} else {
			ret.Key = encodeBytes(msg.Key, flags.EncodeKey)
		}
	}

	return ret, nil
}

func (deserializer ProtobufMessageDeserializer) Deserialize(rawMsg *sarama.ConsumerMessage, flags Flags) error {
	output.Debugf("start to deserialize protobuf message...")

	msg, err := deserializer.newProtobufMessage(rawMsg, flags)
	if err != nil {
		return err
	}

	if flags.OutputFormat != "" {
		return output.PrintObject(msg, flags.OutputFormat)
	}

	var row []string

	if flags.PrintHeaders {
		if msg.Headers != nil {
			column := toSortedArray(msg.Headers)
			row = append(row, strings.Join(column[:], ","))
		} else {
			row = append(row, "")
		}
	}

	if flags.PrintPartitions {
		row = append(row, strconv.Itoa(int(msg.Partition)))
	}

	if flags.PrintKeys {
		if msg.Key != nil {
			row = append(row, *msg.Key)
		} else {
			row = append(row, "")
		}
	}
	if flags.PrintTimestamps {
		if msg.Timestamp != nil {
			row = append(row, (*msg.Timestamp).Format(time.RFC3339))
		} else {
			row = append(row, "")
		}
	}

	var value string

	if msg.Value != nil {
		value = *msg.Value
	} else {
		value = "null"
	}

	row = append(row, value)

	output.PrintStrings(strings.Join(row[:], flags.Separator))
	return nil
}

func (deserializer ProtobufMessageDeserializer) CanDeserialize(_ string) (bool, error) {
	return true, nil
}

func decodeProtobuf(b []byte, msgDesc *desc.MessageDescriptor, encoding string) (*string, error) {
	if len(b) == 0 { // tombstone record, can't be unmarshalled as protobuf
		return nil, nil
	}

	msg := dynamic.NewMessage(msgDesc)
	if err := msg.Unmarshal(b); err != nil {
		return nil, err
	}

	j, err := msg.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return encodeBytes(j, encoding), nil
}
