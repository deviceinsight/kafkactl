package output

import (
	"fmt"
	"os"
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

func (writer *TableWriter) WriteHeader(columns ...string) {
	writer.Initialize()
	_, err := fmt.Fprintln(writer.client, strings.Join(columns[:], "\t"))
	if err != nil {
		Failf("Failed to write table header: %s", err)
	}
}

func (writer *TableWriter) Initialize() {
	writer.client.Init(os.Stdout, 0, 0, 5, ' ', 0)
	writer.initialized = true
}

func (writer *TableWriter) Write(columns ...string) {

	if !writer.initialized {
		Failf("error: no table header written")
	}

	_, err := fmt.Fprintln(writer.client, strings.Join(columns[:], "\t"))
	if err != nil {
		Failf("Failed to write table header: %s", err)
	}
}

func (writer *TableWriter) Flush() {
	err := writer.client.Flush()
	if err != nil {
		Failf("Failed to flush table writer: %s", err)
	}
}
