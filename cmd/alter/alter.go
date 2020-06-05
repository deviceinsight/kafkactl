package alter

import (
	"github.com/spf13/cobra"
)

func NewAlterCmd() *cobra.Command {

	var cmdAlter = &cobra.Command{
		Use:   "alter",
		Short: "alter topics",
	}

	cmdAlter.AddCommand(newAlterTopicCmd())

	return cmdAlter
}
