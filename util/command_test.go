package util

import (
	"github.com/deviceinsight/kafkactl/cmd"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func ExecuteCommandC(args ...string) (c *cobra.Command, out string, err error) {

	streams, buf := output.NewTestIOStreams()

	root := cmd.KafkactlCommand(streams)

	root.SetArgs(args)
	c, err = root.ExecuteC()
	return c, buf.String(), err
}
