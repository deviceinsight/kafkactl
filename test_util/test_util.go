package test_util

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/deviceinsight/kafkactl/cmd"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

var configFile = "it-config.yml"

var testIoStreams output.IOStreams

func getRootDir() (string, error) {

	path, err := os.Getwd()
	if err != nil {
		return "", errors.New("unable to get working dir")
	}

	_, err = os.Stat(filepath.Join(path, configFile))

	for os.IsNotExist(err) {
		if strings.HasSuffix(path, "kafkactl") {
			return "", errors.New("unable to find it-config.yml in root folder")
		}
		oldPath := path

		if path = filepath.Dir(oldPath); path == oldPath {
			return "", errors.New("unable to find it-config.yml")
		}
		_, err = os.Stat(filepath.Join(path, configFile))
	}

	return path, err
}

func init() {

	rootDir, err := getRootDir()
	if err != nil {
		panic(err)
	}

	if err := os.Setenv("KAFKA_CTL_CONFIG", filepath.Join(rootDir, configFile)); err != nil {
		panic(err)
	}
}

func StartUnitTest(t *testing.T) {
	startTest(t, "test.log")
}

func StartIntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	startTest(t, "integration-test.log")
}

func startTest(t *testing.T, logFilename string) {

	rootDir, err := getRootDir()
	if err != nil {
		panic(err)
	}

	logFilename = filepath.Join(rootDir, logFilename)
	logFile, err := os.OpenFile(logFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	testIoStreams = output.NewTestIOStreams(logFile)
	output.TestLogger = log.New(testIoStreams.DebugOut, "[test    ] ", log.LstdFlags)

	output.TestLogf("---")
	output.TestLogf("--- Starting: %s", t.Name())

	t.Cleanup(func() {
		output.TestLogf("---")
		output.TestLogf("--- Finished: %s", t.Name())
		_ = logFile.Close()
	})
}

func AssertEquals(t *testing.T, expected, actual string) {

	if strings.TrimSpace(actual) != strings.TrimSpace(expected) {
		t.Fatalf("unexpected output:\nexpected:\t%s\n  actual:\t%s", expected, strings.TrimSpace(actual))
	}
}

func AssertContains(t *testing.T, expected string, array []string) {
	if !util.ContainsString(array, expected) {
		t.Fatalf("expected array to contain: %s\narray: %v", expected, array)
	}
}

func AssertContainsNot(t *testing.T, unexpected string, array []string) {
	if util.ContainsString(array, unexpected) {
		t.Fatalf("expected array to NOT contain: %s\narray: %v", unexpected, array)
	}
}

func GetPrefixedName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, random.Intn(100000))
}

func WithoutBrokerReferences(output string) string {

	brokerAddressRegex := regexp.MustCompile(`localhost:\d9092`)
	withoutBrokerAddresses := brokerAddressRegex.ReplaceAllString(output, "any-broker")

	brokerIdRegex := regexp.MustCompile(`([^\d\w])(101|102|103)([^\d\w])`)
	return brokerIdRegex.ReplaceAllString(withoutBrokerAddresses, "${1}any-broker-id${3}")
}

type KafkaCtlTestCommand struct {
	Streams output.IOStreams
	Root    *cobra.Command
	Verbose bool
}

func CreateKafkaCtlCommand() (kafkactl KafkaCtlTestCommand) {

	if testIoStreams.Out == nil {
		panic("cannot create CreateKafkaCtlCommand(). Did you call StartUnitTest() or StartIntegrationTest()?")
	}

	return KafkaCtlTestCommand{Streams: testIoStreams, Root: cmd.NewKafkactlCommand(testIoStreams), Verbose: false}
}

func (kafkactl *KafkaCtlTestCommand) Execute(args ...string) (cmd *cobra.Command, err error) {
	// reset output streams
	kafkactl.Streams.Out.(*bytes.Buffer).Reset()
	kafkactl.Streams.ErrOut.(*bytes.Buffer).Reset()

	if kafkactl.Verbose {
		args = append(args, "-V")
	}

	kafkactl.Root.SetArgs(args)

	var specificErr error

	output.Fail = func(err error) {
		specificErr = err
	}

	command, generalErr := kafkactl.Root.ExecuteC()

	output.TestLogf("executed: kafkactl %s", strings.Join(args, " "))
	output.TestLogf("response: %s %s", kafkactl.GetStdOut(), kafkactl.GetStdErr())

	if generalErr != nil {
		return command, generalErr
	} else {
		return command, specificErr
	}
}

func (kafkactl *KafkaCtlTestCommand) GetStdOut() string {
	return kafkactl.Streams.Out.(*bytes.Buffer).String()
}

func (kafkactl *KafkaCtlTestCommand) GetStdErr() string {
	return kafkactl.Streams.ErrOut.(*bytes.Buffer).String()
}
