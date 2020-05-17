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

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

func NewTestIOStreams() (IOStreams, *bytes.Buffer) {
	return NewTestIOStreamsWithStdIn(nil)
}

func NewTestIOStreamsWithStdIn(in *os.File) (IOStreams, *bytes.Buffer) {

	buf := new(bytes.Buffer)

	return IOStreams{
		In:     in,
		Out:    buf,
		ErrOut: buf,
	}, buf
}
