package output

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"text/tabwriter"
)

type TableWriter struct {
	client      *tabwriter.Writer
	initialized bool
}

func CreateTableWriter() TableWriter {

	var writer TableWriter

	writer.client = new(tabwriter.Writer)
	writer.initialized = false

	return writer
}

func (writer *TableWriter) WriteHeader(columns ...string) error {
	writer.Initialize()
	_, err := fmt.Fprintln(writer.client, strings.Join(columns[:], "\t"))

	if err != nil {
		return errors.Wrap(err, "Failed to write table header")
	} else {
		return nil
	}
}

func (writer *TableWriter) Initialize() {
	writer.client.Init(IoStreams.Out, 0, 0, 5, ' ', 0)
	writer.initialized = true
}

func (writer *TableWriter) Write(columns ...string) error {

	if !writer.initialized {
		return errors.New("no table header written")
	}

	_, err := fmt.Fprintln(writer.client, strings.Join(columns[:], "\t"))
	if err != nil {
		return errors.Wrap(err, "Failed to write table header")
	} else {
		return nil
	}
}

func (writer *TableWriter) Flush() error {
	err := writer.client.Flush()
	if err != nil {
		return errors.Wrap(err, "Failed to flush table writer")
	} else {
		return nil
	}
}
