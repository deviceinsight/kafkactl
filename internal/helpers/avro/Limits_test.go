package avro

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/linkedin/goavro/v2"
)

func TestGoAvroMaxBlockCountIsConfigured(t *testing.T) {
	const wantMaxBlockCount int64 = 1_000_000
	if maxAvroBlockCount != wantMaxBlockCount {
		t.Fatalf("maxAvroBlockCount = %d, want %d", maxAvroBlockCount, wantMaxBlockCount)
	}
	if goavro.MaxBlockCount != wantMaxBlockCount {
		t.Fatalf("goavro.MaxBlockCount = %d, want %d", goavro.MaxBlockCount, wantMaxBlockCount)
	}
}

func TestMessageCodecRejectsBlocksOverLimit(t *testing.T) {
	if goavro.MaxBlockCount != maxAvroBlockCount {
		t.Fatalf("goavro.MaxBlockCount = %d, want %d", goavro.MaxBlockCount, maxAvroBlockCount)
	}

	tests := []struct {
		name   string
		schema string
	}{
		{name: "array", schema: `{"type":"array","items":"null"}`},
		{name: "map", schema: `{"type":"map","values":"null"}`},
	}

	var encodedBlockCount [binary.MaxVarintLen64]byte
	encodedLength := binary.PutUvarint(
		encodedBlockCount[:],
		uint64(maxAvroBlockCount+1)<<1,
	)
	payload := encodedBlockCount[:encodedLength]

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			codec, err := NewMessageCodec(test.schema, Standard)
			if err != nil {
				t.Fatalf("NewMessageCodec(): unexpected error: %v", err)
			}

			_, err = codec.DecodeBinary(payload)
			if err == nil {
				t.Fatal("DecodeBinary() returned nil error for a block count above the limit")
			}
			if !strings.Contains(err.Error(), "block count exceeds MaxBlockCount") {
				t.Fatalf(
					"DecodeBinary() error = %q, want it to contain %q",
					err,
					"block count exceeds MaxBlockCount",
				)
			}
		})
	}
}
