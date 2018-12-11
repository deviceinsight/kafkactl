package create

import (
	"github.com/spf13/cobra"
)

var CmdCreate = &cobra.Command{
	Use:   "create",
	Short: "create topics",
}

func init() {
	CmdCreate.AddCommand(cmdCreateTopic)
}
