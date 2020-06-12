package config

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCurrentContextCmd() *cobra.Command {
	var cmdCurrentContext = &cobra.Command{
		Use:     "current-context",
		Aliases: []string{"currentContext"},
		Short:   "show current context",
		Long:    `Displays the name of context that is currently active`,
		Run: func(cmd *cobra.Command, args []string) {
			context := viper.GetString("current-context")
			output.Infof("%s", context)
		},
	}

	return cmdCurrentContext
}
