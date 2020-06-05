package get

import (
	"github.com/spf13/cobra"
)

func NewGetCmd() *cobra.Command {

	var cmdGet = &cobra.Command{
		Use:     "get",
		Aliases: []string{"list"},
		Short:   "get info about topics, consumerGroups",
	}

	cmdGet.AddCommand(newGetTopicsCmd())
	cmdGet.AddCommand(newGetConsumerGroupsCmd())

	return cmdGet
}
