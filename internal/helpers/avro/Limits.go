package avro

import "github.com/linkedin/goavro/v2"

// maxAvroBlockCount limits the number of items in a single Avro array or map
// block. It is not a cumulative limit across blocks or nested collections.
const maxAvroBlockCount int64 = 1_000_000

func init() {
	// Set the process-wide limit during package initialization so it is not
	// mutated while codecs are in use.
	goavro.MaxBlockCount = maxAvroBlockCount
}
