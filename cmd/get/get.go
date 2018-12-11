package get

import (
	"github.com/spf13/cobra"
)

var CmdGet = &cobra.Command{
	Use:   "get",
	Short: "get info about topics",
}

func init() {
	CmdGet.AddCommand(cmdGetTopics)
}
