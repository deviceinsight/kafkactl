package cloudevents

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"io"

	"github.com/pkg/errors"
)

type jsonEventMember struct {
	name  string
	value json.RawMessage
}

// JSONEvent keeps the members of a JSON format CloudEvent raw and in order, so re-encoding only changes replaced ones.
type JSONEvent struct {
	members []jsonEventMember
}

func ParseJSONEvent(value []byte) (*JSONEvent, error) {

	decoder := json.NewDecoder(bytes.NewReader(value))

	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return nil, errors.New("event is not a json object")
	}

	event := &JSONEvent{}

	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		name, ok := token.(string)
		if !ok {
			return nil, errors.Errorf("expected a member name, got %v", token)
		}

		member := jsonEventMember{name: name}

		if err := decoder.Decode(&member.value); err != nil {
			return nil, err
		}

		event.members = append(event.members, member)
	}

	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, errors.New("event is not a json object")
	}

	// only a clean EOF means the object was the whole value. Anything else is trailing data,
	// whether it parses as a further token or is malformed, and must not be silently dropped.
	if _, err := decoder.Token(); !stderrors.Is(err, io.EOF) {
		return nil, errors.New("unexpected data after the event json object")
	}

	return event, nil
}

func (event *JSONEvent) Get(name string) (json.RawMessage, bool) {
	for _, member := range event.members {
		if member.name == name {
			return member.value, true
		}
	}
	return nil, false
}

func (event *JSONEvent) Replace(oldName, newName string, value json.RawMessage) {
	for i, member := range event.members {
		if member.name == oldName {
			event.members[i] = jsonEventMember{name: newName, value: value}
			return
		}
	}
}

func (event *JSONEvent) Marshal() []byte {
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
