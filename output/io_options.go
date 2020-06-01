package output

import (
	"bytes"
	"io"
	"os"
	"testing"
)

var IoStreams = DefaultIOStreams()

func DefaultIOStreams() IOStreams {
	return IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
}

func NewTestIOStreams(t *testing.T) IOStreams {
	return NewTestIOStreamsWithStdIn(t, nil)
}

func NewTestIOStreamsWithStdIn(t *testing.T, in *os.File) IOStreams {

	streams := IOStreams{
		In:     in,
		Out:    new(bytes.Buffer),
		ErrOut: new(bytes.Buffer),
	}

	IoStreams = streams
	Fail = func(err error) {
		t.Fatalf(err.Error())
	}

	return streams
}

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}
