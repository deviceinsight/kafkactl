package create

import "github.com/spf13/cobra"

func NewCreateCmd() *cobra.Command {

	var cmdCreate = &cobra.Command{
		Use:   "create",
		Short: "create topics, consumerGroups, acls",
	}

	cmdCreate.AddCommand(newCreateTopicCmd())
	cmdCreate.AddCommand(newCreateConsumerGroupCmd())
	cmdCreate.AddCommand(newCreateACLCmd())
	return cmdCreate
}
