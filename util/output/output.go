package output

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

func Quitf(msg string, args ...interface{}) {
	Exitf(0, msg, args...)
}

func Failf(msg string, args ...interface{}) {
	Exitf(1, msg, args...)
}

func Infof(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, msg+"\n", args...)
}

func Exitf(code int, msg string, args ...interface{}) {
	if code == 0 {
		fmt.Fprintf(os.Stdout, msg+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, msg+"\n", args...)
	}
	os.Exit(code)
}

func PrintObject(object interface{}, format string) {
	if format == "yaml" {
		yamlString, err := yaml.Marshal(object)
		if err != nil {
			Failf("unable to format yaml: %v", err)
		}
		fmt.Println(string(yamlString))
	} else if format == "json" {
		jsonString, err := json.MarshalIndent(object, "", "\t")
		if err != nil {
			Failf("unable to format json: %v", err)
		}
		fmt.Println(string(jsonString))
	} else {
		Failf("unknown format: %v", format)
	}
}

func PrintStrings(args ...string) {
	for _, arg := range args {
		fmt.Println(arg)
	}
}
