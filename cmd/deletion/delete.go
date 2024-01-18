package deletion

import (
	"github.com/spf13/cobra"
)

func NewDeleteCmd() *cobra.Command {

	var cmdDelete = &cobra.Command{
		Use:   "delete",
		Short: "delete topics, consumerGroups, consumer-group-offset, acls, records",
	}

	cmdDelete.AddCommand(newDeleteTopicCmd())
	cmdDelete.AddCommand(newDeleteConsumerGroupCmd())
	cmdDelete.AddCommand(newDeleteConsumerGroupOffsetCmd())
	cmdDelete.AddCommand(newDeleteACLCmd())
	cmdDelete.AddCommand(newDeleteRecordsCmd())
	return cmdDelete
}
