package output

import (
	"bytes"
	"io"
	"io/ioutil"
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

func NewTestIOStreams() (IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	return IOStreams{
		In:     in,
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}

func NewTestIOStreamsDiscard() IOStreams {
	in := &bytes.Buffer{}
	return IOStreams{
		In:     in,
		Out:    ioutil.Discard,
		ErrOut: ioutil.Discard,
	}
}
