package output

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/IBM/sarama"
)

var IoStreams = DefaultIOStreams()

func DefaultIOStreams() IOStreams {
	return IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr, DebugOut: os.Stderr}
}

func NewTestIOStreams(debug *os.File) IOStreams {
	return NewTestIOStreamsWithStdIn(nil, debug)
}

func NewTestIOStreamsWithStdIn(in *os.File, debug *os.File) IOStreams {

	streams := IOStreams{
		In:       in,
		Out:      new(bytes.Buffer),
		ErrOut:   new(bytes.Buffer),
		DebugOut: debug,
	}

	IoStreams = streams

	return streams
}

type IOStreams struct {
	In       io.Reader
	Out      io.Writer
	ErrOut   io.Writer
	DebugOut io.Writer
}

func (streams *IOStreams) EnableDebug() {
	sarama.Logger = log.New(streams.DebugOut, "[sarama  ] ", log.LstdFlags)
	DebugLogger = log.New(streams.DebugOut, "[kafkactl] ", log.LstdFlags)
}
