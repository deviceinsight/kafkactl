package cmd_test

import (
	"strings"
	"testing"

	"github.com/deviceinsight/kafkactl/v5/cmd"
	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestVersionCommand(t *testing.T) {
	t.Parallel()

	testutil.StartUnitTest(t)

	cmd.Version = "1.8.0"
	cmd.GitCommit = "ef6a0263c9623d44d198a0f39d712ddb76bb5c04"
	cmd.BuildTime = "2020-05-21T11:14:58+00:00"

	kafkactl := testutil.CreateKafkaCtlCommand()

	if _, err := kafkactl.Execute("version"); err != nil {
		t.Fatalf("failed to execute version command: %v", err)
	}

	expected := "cmd.info{version:\"1.8.0\", buildTime:\"2020-05-21T11:14:58+00:00\", gitCommit:\"ef6a0263c9623d44d198a0f39d712ddb76bb5c04\""

	if !strings.HasPrefix(kafkactl.GetStdOut(), expected) {
		t.Fatalf("unexpected output: %s", kafkactl.GetStdOut())
	}
}
