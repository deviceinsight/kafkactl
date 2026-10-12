package consume

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

type jsonEventMember struct {
	name  string
	value json.RawMessage
}

// jsonEvent keeps the members of a JSON format CloudEvent raw and in order, so re-encoding only changes replaced ones.
type jsonEvent struct {
	members []jsonEventMember
}

func parseJSONEvent(value []byte) (*jsonEvent, error) {
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()

	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return nil, errors.New("event is not a json object")
	}

	event := &jsonEvent{}

	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		var member jsonEventMember
		member.name = token.(string)

		if err := decoder.Decode(&member.value); err != nil {
			return nil, err
		}

		event.members = append(event.members, member)
	}

	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, errors.New("event is not a json object")
	}

	if _, err := decoder.Token(); err == nil {
		return nil, errors.New("unexpected data after the event json object")
	}

	return event, nil
}

func (event *jsonEvent) get(name string) (json.RawMessage, bool) {
	for _, member := range event.members {
		if member.name == name {
			return member.value, true
		}
	}
	return nil, false
}

func (event *jsonEvent) replace(oldName, newName string, value json.RawMessage) {
	for i, member := range event.members {
		if member.name == oldName {
			event.members[i] = jsonEventMember{name: newName, value: value}
			return
		}
	}
}

func (event *jsonEvent) marshal() []byte {
	var buffer bytes.Buffer

	buffer.WriteByte('{')
	for i, member := range event.members {
		if i > 0 {
			buffer.WriteByte(',')
		}
		name, _ := json.Marshal(member.name)
		buffer.Write(name)
		buffer.WriteByte(':')
		buffer.Write(member.value)
	}
	buffer.WriteByte('}')

	return buffer.Bytes()
}
