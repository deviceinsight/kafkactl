package alter

import (
	"github.com/spf13/cobra"
)

func NewAlterCmd() *cobra.Command {

	var cmdAlter = &cobra.Command{
		Use:     "alter",
		Aliases: []string{"edit"},
		Short:   "alter topics, partitions, brokers, users",
	}

	cmdAlter.AddCommand(newAlterTopicCmd())
	cmdAlter.AddCommand(newAlterPartitionCmd())
	cmdAlter.AddCommand(newAlterBrokerCmd())
	cmdAlter.AddCommand(newAlterUserCmd())
	return cmdAlter
}
