package input

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/deviceinsight/kafkactl/v5/internal/util"
)

const defaultKeyColumnIdx = 0
const defaultValueColumnIdx = 1
const defaultColumnCount = 2

type csvParser struct {
	key            string
	separator      string
	keyColumnIdx   int
	valueColumnIdx int
	columnCount    int
}

func NewCsvParser(key string, separator string) Parser {
	return &csvParser{
		key:            key,
		separator:      separator,
		keyColumnIdx:   defaultKeyColumnIdx,
		valueColumnIdx: defaultValueColumnIdx,
		columnCount:    defaultColumnCount,
	}
}

func (p *csvParser) ParseLine(line string) (Message, error) {

	if p.separator == "" {
		return Message{Key: &p.key, Value: &line}, nil
	}

	input := strings.Split(line, util.ConvertControlChars(p.separator))
	if len(input) < 2 {
		return Message{}, fmt.Errorf("the provided input does not contain the separator %s", p.separator)
	} else if len(input) == 3 && p.columnCount == defaultColumnCount {
		// lazy resolving of column indices
		var err error
		p.keyColumnIdx, p.valueColumnIdx, p.columnCount, err = resolveColumns(input)
		if err != nil {
			return Message{}, err
		}
	} else if len(input) != p.columnCount {
		return Message{}, fmt.Errorf("line contains unexpected amount of separators:\n%s", line)
	}

	return Message{&input[p.keyColumnIdx], &input[p.valueColumnIdx], nil}, nil
}

func resolveColumns(line []string) (keyColumnIdx, valueColumnIdx, columnCount int, err error) {
	if isTimestamp(line[0]) {
		output.Warnf("assuming column 0 to be message timestamp. Column will be ignored")
		return 1, 2, 3, nil
	} else if isTimestamp(line[1]) {
		output.Warnf("assuming column 1 to be message timestamp. Column will be ignored")
		return 0, 2, 3, nil
	}
	return -1, -1, -1, errors.Errorf("line contains unexpected amount of separators:\n%s", line)
}

func isTimestamp(value string) bool {
	_, e := time.Parse(time.RFC3339, value)
	return e == nil
}
