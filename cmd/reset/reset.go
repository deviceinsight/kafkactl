package reset

import (
	"github.com/spf13/cobra"
)

var CmdReset = &cobra.Command{
	Use:   "reset",
	Short: "reset consumerGroupsOffset",
}

func init() {
	CmdReset.AddCommand(cmdResetOffset)
}
