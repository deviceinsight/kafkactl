package test_util

import (
	"bytes"
	"fmt"
	"github.com/deviceinsight/kafkactl/cmd"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func init() {

	configFile := "it-config.yml"

	path, err := os.Getwd()
	if err != nil {
		panic("unable to get working dir")
	}

	_, err = os.Stat(filepath.Join(path, configFile))

	for os.IsNotExist(err) {
		if strings.HasSuffix(path, "kafkactl") {
			panic("unable to find it-config.yml in root folder")
		}
		oldPath := path

		if path = filepath.Dir(oldPath); path == oldPath {
			panic("unable to find it-config.yml")
		}
		_, err = os.Stat(filepath.Join(path, configFile))
	}

	if err := os.Setenv("KAFKA_CTL_CONFIG", filepath.Join(path, configFile)); err != nil {
		panic(err)
	}
}

func AssertEquals(t *testing.T, expected, actual string) {

	if strings.TrimSpace(actual) != strings.TrimSpace(expected) {
		t.Fatalf("unexpected output:\nexpected:\t%s\n  actual:\t%s", expected, strings.TrimSpace(actual))
	}
}

func GetTopicName(prefix string) string {
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
}

func CreateKafkaCtlCommand() (kafkactl KafkaCtlTestCommand) {
	streams := output.NewTestIOStreams()
	return KafkaCtlTestCommand{Streams: streams, Root: cmd.NewKafkactlCommand(streams)}
}

func (kafkactl *KafkaCtlTestCommand) Execute(args ...string) (cmd *cobra.Command, err error) {
	// reset output streams
	kafkactl.Streams.Out.(*bytes.Buffer).Reset()
	kafkactl.Streams.ErrOut.(*bytes.Buffer).Reset()

	kafkactl.Root.SetArgs(args)

	var specificErr error

	output.Fail = func(err error) {
		specificErr = err
	}

	command, generalErr := kafkactl.Root.ExecuteC()

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
