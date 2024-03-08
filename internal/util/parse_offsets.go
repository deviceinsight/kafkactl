package util

import (
	"math"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const ErrOffset = math.MinInt64
const offsetSeparator = "="

func ParseOffsets(rawOffsets []string) (map[int32]int64, error) {

	offsets := make(map[int32]int64)

	for _, offsetFlag := range rawOffsets {
		offsetParts := strings.Split(offsetFlag, offsetSeparator)

		if len(offsetParts) != 2 {
			return nil, errors.Errorf("offset parameter has wrong format: %s %v", offsetFlag, rawOffsets)
		}

		partition, err := strconv.Atoi(offsetParts[0])
		if err != nil {
			return nil, errors.Errorf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
		}

		offset, err := strconv.ParseInt(offsetParts[1], 10, 64)
		if err != nil {
			return nil, errors.Errorf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
		}

		offsets[int32(partition)] = offset
	}

	return offsets, nil
}

func ExtractOffsetForPartition(rawOffsets []string, currentPartition int32) (int64, error) {
	for _, offsetFlag := range rawOffsets {
		offsetParts := strings.Split(offsetFlag, offsetSeparator)

		if len(offsetParts) == 2 {

			partition, err := strconv.Atoi(offsetParts[0])
			if err != nil {
				return ErrOffset, errors.Errorf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
			}

			if int32(partition) != currentPartition {
				continue
			}

			offset, err := strconv.ParseInt(offsetParts[1], 10, 64)
			if err != nil {
				return ErrOffset, errors.Errorf("unable to parse offset parameter: %s (%v)", offsetFlag, err)
			}

			return offset, nil
		}
	}
	return ErrOffset, errors.Errorf("unable to find offset parameter for partition %d: %v", currentPartition, rawOffsets)
}
