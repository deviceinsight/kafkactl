package create

import "github.com/spf13/cobra"

func NewCreateCmd() *cobra.Command {

	var cmdCreate = &cobra.Command{
		Use:   "create",
		Short: "create topics, consumerGroups, acls, users",
	}

	cmdCreate.AddCommand(newCreateTopicCmd())
	cmdCreate.AddCommand(newCreateConsumerGroupCmd())
	cmdCreate.AddCommand(newCreateACLCmd())
	cmdCreate.AddCommand(newCreateUserCmd())
	return cmdCreate
}
