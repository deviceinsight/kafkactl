package input

type Message struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Parser interface {
	ParseLine(line string) (Message, error)
}
