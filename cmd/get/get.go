package get

import (
	"github.com/spf13/cobra"
)

var CmdGet = &cobra.Command{
	Use:     "get",
	Aliases: []string{"list"},
	Short:   "get info about topics, consumerGroups",
}

func init() {
	CmdGet.AddCommand(cmdGetTopics)
	CmdGet.AddCommand(cmdGetConsumerGroups)
}
