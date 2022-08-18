package clone

import "github.com/spf13/cobra"

func NewCloneCmd() *cobra.Command {

	var cmdClone = &cobra.Command{
		Use:   "clone",
		Short: "clone topics, consumerGroups",
	}

	cmdClone.AddCommand(newCloneTopicCmd())
	cmdClone.AddCommand(newCloneConsumerGroupCmd())

	return cmdClone
}
