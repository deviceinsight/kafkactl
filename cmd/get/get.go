package get

import (
	"github.com/spf13/cobra"
)

func NewGetCmd() *cobra.Command {

	var cmdGet = &cobra.Command{
		Use:     "get",
		Aliases: []string{"list"},
		Short:   "get info about topics, consumerGroups, acls, brokers",
	}

	cmdGet.AddCommand(newGetTopicsCmd())
	cmdGet.AddCommand(newGetConsumerGroupsCmd())
	cmdGet.AddCommand(newGetACLCmd())
	cmdGet.AddCommand(newGetBrokersCmd())

	return cmdGet
}
