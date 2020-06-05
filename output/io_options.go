package output

import (
	"bytes"
	"io"
	"os"
)

var IoStreams = DefaultIOStreams()

func DefaultIOStreams() IOStreams {
	return IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
}

func NewTestIOStreams() IOStreams {
	return NewTestIOStreamsWithStdIn(nil)
}

func NewTestIOStreamsWithStdIn(in *os.File) IOStreams {

	streams := IOStreams{
		In:     in,
		Out:    new(bytes.Buffer),
		ErrOut: new(bytes.Buffer),
	}

	IoStreams = streams

	return streams
}

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}
