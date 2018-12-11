package deletion

import (
	"github.com/spf13/cobra"
)

var CmdDelete = &cobra.Command{
	Use:   "delete",
	Short: "delete topics",
}

func init() {
	CmdDelete.AddCommand(cmdDeleteTopic)
}
