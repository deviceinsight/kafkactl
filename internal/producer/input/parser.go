package input

type Message struct {
	Key     *string           `json:"key"`
	Value   *string           `json:"value"`
	Headers map[string]string `json:"headers,omitempty"`
}

type Parser interface {
	ParseLine(line string) (Message, error)
}
