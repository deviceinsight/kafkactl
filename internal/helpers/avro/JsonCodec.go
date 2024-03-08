package avro

import (
	"strings"

	"github.com/deviceinsight/kafkactl/internal/output"
)

type JSONCodec int

const (
	Standard JSONCodec = iota
	Avro
)

func (codec JSONCodec) String() string {
	return []string{"standard", "avro"}[codec]
}

func ParseJSONCodec(codec string) JSONCodec {

	switch strings.ToLower(codec) {
	case "":
		fallthrough
	case "standard":
		return Standard
	case "avro":
		return Avro
	default:
		output.Warnf("unable to parse avro json codec: %s", codec)
		return Standard
	}
}
