package deletion

import (
	"github.com/spf13/cobra"
)

func NewDeleteCmd() *cobra.Command {

	var cmdDelete = &cobra.Command{
		Use:   "delete",
		Short: "delete topics, acls",
	}

	cmdDelete.AddCommand(newDeleteTopicCmd())
	cmdDelete.AddCommand(newDeleteAclCmd())
	return cmdDelete
}
