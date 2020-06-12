package reset

import (
	"github.com/spf13/cobra"
)

func NewResetCmd() *cobra.Command {
	var cmdReset = &cobra.Command{
		Use:   "reset",
		Short: "reset consumerGroupsOffset",
	}

	cmdReset.AddCommand(newResetOffsetCmd())

	return cmdReset
}
