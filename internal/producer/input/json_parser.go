package input

import (
	"encoding/json"
	"fmt"
)

type jsonParser struct {
}

func NewJSONParser() Parser {
	return &jsonParser{}
}

func (p *jsonParser) ParseLine(line string) (Message, error) {

	var message Message

	if err := json.Unmarshal([]byte(line), &message); err != nil {
		return message, fmt.Errorf("can't unmarshal line: %w", err)
	}

	return message, nil
}
