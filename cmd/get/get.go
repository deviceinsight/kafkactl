package get

import (
	"github.com/spf13/cobra"
)

func NewGetCmd() *cobra.Command {

	var cmdGet = &cobra.Command{
		Use:     "get",
		Aliases: []string{"list"},
		Short:   "get info about topics, consumerGroups, acls",
	}

	cmdGet.AddCommand(newGetTopicsCmd())
	cmdGet.AddCommand(newGetConsumerGroupsCmd())
	cmdGet.AddCommand(newGetAclCmd())

	return cmdGet
}
