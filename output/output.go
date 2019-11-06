package output

import (
	"encoding/json"
	"fmt"
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

func Failf(msg string, args ...interface{}) {
	Exitf(1, msg, args...)
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

func Exitf(code int, msg string, args ...interface{}) {
	if code == 0 {
		_, _ = fmt.Fprintf(IoStreams.Out, msg+"\n", args...)
	} else {
		_, _ = fmt.Fprintf(IoStreams.ErrOut, msg+"\n", args...)
	}
	os.Exit(code)
}

func PrintObject(object interface{}, format string) {
	if format == "yaml" {
		yamlString, err := yaml.Marshal(object)
		if err != nil {
			Failf("unable to format yaml: %v", err)
		}
		_, _ = fmt.Fprintln(IoStreams.Out, string(yamlString))
	} else if format == "json" {
		jsonString, err := json.MarshalIndent(object, "", "\t")
		if err != nil {
			Failf("unable to format json: %v", err)
		}
		_, _ = fmt.Fprintln(IoStreams.Out, string(jsonString))
	} else {
		Failf("unknown format: %v", format)
	}
}

func PrintStrings(args ...string) {
	for _, arg := range args {
		_, _ = fmt.Fprintln(IoStreams.Out, arg)
	}
}
