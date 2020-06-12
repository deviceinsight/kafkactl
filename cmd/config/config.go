package config

import (
	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {

	var cmdConfig = &cobra.Command{
		Use:   "config",
		Short: "show and edit configurations",
	}

	cmdConfig.AddCommand(newCurrentContextCmd())
	cmdConfig.AddCommand(newGetContextsCmd())
	cmdConfig.AddCommand(newUseContextCmd())
	cmdConfig.AddCommand(newViewCmd())

	return cmdConfig
}
