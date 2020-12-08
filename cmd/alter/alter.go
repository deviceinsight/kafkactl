package alter

import (
	"github.com/spf13/cobra"
)

func NewAlterCmd() *cobra.Command {

	var cmdAlter = &cobra.Command{
		Use:     "alter",
		Aliases: []string{"edit"},
		Short:   "alter topics, partitions",
	}

	cmdAlter.AddCommand(newAlterTopicCmd())
	cmdAlter.AddCommand(newAlterPartitionCmd())
	return cmdAlter
}
