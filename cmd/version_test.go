package cmd

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	t.Parallel()

	version = "1.8.0"
	gitCommit = "ef6a0263c9623d44d198a0f39d712ddb76bb5c04"
	buildTime = "2020-05-21T11:14:58+00:00"

	_, out, err := executeCommandC("version")

	if err != nil {
		t.Fatalf("failed to execute version command: %v", err)
	}

	expected := "cmd.info{version:\"1.8.0\", buildTime:\"2020-05-21T11:14:58+00:00\", gitCommit:\"ef6a0263c9623d44d198a0f39d712ddb76bb5c04\""

	if !strings.HasPrefix(out, expected) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func executeCommandC(args ...string) (c *cobra.Command, out string, err error) {

	streams, buf := output.NewTestIOStreams()

	root := KafkactlCommand(streams)

	root.SetArgs(args)
	c, err = root.ExecuteC()
	return c, buf.String(), err
}
