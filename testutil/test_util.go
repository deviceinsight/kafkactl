package testutil

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/deviceinsight/kafkactl/internal/env"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/cmd"
	"github.com/deviceinsight/kafkactl/internal"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/deviceinsight/kafkactl/util"
	"github.com/spf13/cobra"
)

var RootDir string

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

	RootDir = rootDir

	if err := os.Setenv("KAFKA_CTL_CONFIG", filepath.Join(rootDir, configFile)); err != nil {
		panic(err)
	}

	for _, variable := range env.Variables {
		if err := os.Setenv(variable, ""); err != nil {
			panic(err)
		}
	}
}

func StartUnitTest(t *testing.T) {
	startTest(t, "test.log")
}

func StartIntegrationTest(t *testing.T) {
	StartIntegrationTestWithContext(t, "default")
}

func StartIntegrationTestWithContext(t *testing.T, context string) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	SwitchContext(context)

	startTest(t, "integration-test.log")
}

func CreateClient(t *testing.T) sarama.Client {
	var (
		err     error
		context internal.ClientContext
		client  sarama.Client
	)

	if context, err = internal.CreateClientContext(); err != nil {
		t.Fatalf("failed to create context : %s", err)
	}

	if client, err = internal.CreateClient(&context); err != nil {
		t.Fatalf("failed to create cluster admin : %s", err)
	}
	return client
}

func MarkOffset(t *testing.T, client sarama.Client, groupName string, topic string, partition int32, offset int64) {
	offsetMgr, _ := sarama.NewOffsetManagerFromClient(groupName, client)
	defer func(offsetMgr sarama.OffsetManager) {
		err := offsetMgr.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(offsetMgr)

	partitionOffsetManager, err := offsetMgr.ManagePartition(topic, partition)
	if err != nil {
		t.Fatal(err)
	}

	defer func(partitionOffsetManager sarama.PartitionOffsetManager) {
		err := partitionOffsetManager.Close()
		if err != nil {
			t.Fatal(err)
		}
	}(partitionOffsetManager)

	partitionOffsetManager.MarkOffset(offset, "")
	offsetMgr.Commit()
}

func SwitchContext(context string) {
	if err := os.Setenv("CURRENT_CONTEXT", context); err != nil {
		panic(err)
	}
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
		t.Fatalf("unexpected output:\nexpected:\n--\n%s\n--\nactual:\n--\n%s\n--", expected, strings.TrimSpace(actual))
	}
}

func AssertArraysEquals(t *testing.T, expected, actual []string) {
	sort.Strings(expected)
	sort.Strings(actual)

	if !util.StringArraysEqual(actual, expected) {
		t.Fatalf("unexpected values:\nexpected:\n--\n%s\n--\nactual:\n--\n%s\n--", expected, actual)
	}
}

func AssertErrorContainsOneOf(t *testing.T, expected []string, err error) {

	if err == nil {
		t.Fatalf("expected error to contain: %s\n: %v", expected, "nil")
	}

	for _, expect := range expected {
		if strings.Contains(err.Error(), expect) {
			return
		}
	}

	t.Fatalf("expected error to contain one of: %s\n: %v", expected, err)
}

func AssertErrorContains(t *testing.T, expected string, err error) {

	if err == nil {
		t.Fatalf("expected error to contain: %s\n: %v", expected, "nil")
	}

	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain: %s\n: %v", expected, err)
	}
}

func AssertContainSubstring(t *testing.T, expected, actual string) {
	if !strings.Contains(actual, expected) {
		t.Fatalf("expected string to contain: %s actual: %s", expected, actual)
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

	brokerAddressRegex := regexp.MustCompile(`localhost:\d909[2|3]`)
	withoutBrokerAddresses := brokerAddressRegex.ReplaceAllString(output, "any-broker")

	brokerIDRegex := regexp.MustCompile(`([^\d\w])(101|102|103)([^\d\w])`)
	return brokerIDRegex.ReplaceAllString(withoutBrokerAddresses, "${1}any-broker-id${3}")
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

	return KafkaCtlTestCommand{Streams: testIoStreams, Root: cmd.NewKafkactlCommand(testIoStreams), Verbose: true}
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
	}
	return command, specificErr
}

func (kafkactl *KafkaCtlTestCommand) GetStdOut() string {
	return kafkactl.Streams.Out.(*bytes.Buffer).String()
}

func (kafkactl *KafkaCtlTestCommand) GetStdOutLines() []string {

	space := regexp.MustCompile(`[[:blank:]]{2,}`)

	stdOutput := space.ReplaceAllString(kafkactl.GetStdOut(), "|")

	return strings.Split(stdOutput, "\n")
}

func (kafkactl *KafkaCtlTestCommand) GetStdErr() string {
	return kafkactl.Streams.ErrOut.(*bytes.Buffer).String()
}
