package output

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

var DebugLogger StdLogger = log.New(ioutil.Discard, "[kafkactl] ", log.LstdFlags)

// StdLogger is used to log error messages.
type StdLogger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

var Fail = func(err error) {
	_, _ = fmt.Fprintf(IoStreams.ErrOut, "%s", err.Error())
	os.Exit(1)
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
	} else {
		return errors.Errorf("unknown format: %v", format)
	}
	return nil
}

func PrintStrings(args ...string) {
	for _, arg := range args {
		_, _ = fmt.Fprintln(IoStreams.Out, arg)
	}
}
