package config

import (
	"github.com/spf13/cobra"
)

var CmdConfig = &cobra.Command{
	Use:   "config",
	Short: "show and edit configurations",
}

func init() {
	CmdConfig.AddCommand(cmdCurrentContext)
	CmdConfig.AddCommand(cmdGetContexts)
	CmdConfig.AddCommand(cmdUseContext)
}
