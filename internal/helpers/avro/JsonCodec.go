package avro

import (
	"strings"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/linkedin/goavro/v2"
)

type JSONCodec int

const (
	Standard JSONCodec = iota
	Avro
)

func (codec JSONCodec) String() string {
	return []string{"standard", "avro"}[codec]
}

// CompileCodec compiles an avro schema into a goavro.Codec using the given JSONCodec's
// textual representation (standard JSON, or Avro's own JSON variant).
func CompileCodec(schema string, jsonCodec JSONCodec) (*goavro.Codec, error) {
	if jsonCodec == Avro {
		return goavro.NewCodec(schema)
	}
	return goavro.NewCodecForStandardJSONFull(schema)
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
