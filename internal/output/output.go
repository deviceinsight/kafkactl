package output

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var DebugLogger StdLogger = CreateLogger(io.Discard, "kafkactl")
var TestLogger StdLogger = CreateLogger(io.Discard, "test")

// StdLogger is used to log error messages.
type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

func Warnf(msg string, args ...interface{}) {
	_, _ = fmt.Fprintf(IoStreams.ErrOut, msg+"\n", args...)
}

func Infof(msg string, args ...interface{}) {
	_, _ = fmt.Fprintf(IoStreams.Out, msg+"\n", args...)
}

func Statusf(msg string, args ...interface{}) {
	_, _ = fmt.Fprintf(IoStreams.Out, msg, args...)
}

func Debugf(msg string, args ...interface{}) {
	DebugLogger.Printf(msg+"\n", args...)
}

func TestLogf(msg string, args ...interface{}) {
	TestLogger.Printf(msg+"\n", args...)
}

func PrintObject(object interface{}, format string) error {
	if format == "yaml" {
		yamlString, err := yaml.Marshal(object)
		if err != nil {
			return errors.Wrap(err, "unable to format yaml")
		}
		_, _ = fmt.Fprintln(IoStreams.Out, string(yamlString))
	} else if format == "json" {
		jsonString, err := json.MarshalIndent(object, "", "\t")
		if err != nil {
			return errors.Wrap(err, "unable to format json")
		}
		_, _ = fmt.Fprintln(IoStreams.Out, string(jsonString))
	} else if format == "json-raw" {
		jsonString, err := json.Marshal(object)
		if err != nil {
			return errors.Wrap(err, "unable to format json")
		}
		_, _ = fmt.Fprintln(IoStreams.Out, string(jsonString))
	} else if format != "none" {
		return errors.Errorf("unknown format: %v", format)
	}
	return nil
}

func PrintStrings(args ...string) {
	for _, arg := range args {
		_, _ = fmt.Fprintln(IoStreams.Out, arg)
	}
}

func CreateLogger(out io.Writer, prefix string) *log.Logger {

	// make sure prefix has a fixed length
	prefix = fmt.Sprintf("%-8s", prefix)[0:8]

	return log.New(out, fmt.Sprintf("[%s] ", prefix), log.LstdFlags)
}

func CreateVerboseLogger(prefix string, verbose bool) *log.Logger {
	logWriter := io.Discard
	if verbose {
		logWriter = os.Stderr
	}
	return CreateLogger(logWriter, prefix)
}
