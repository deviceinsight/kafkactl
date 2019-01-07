package alter

import (
	"github.com/spf13/cobra"
)

var CmdAlter = &cobra.Command{
	Use:   "alter",
	Short: "alter topics",
}

func init() {
	CmdAlter.AddCommand(cmdAlterTopic)
}
