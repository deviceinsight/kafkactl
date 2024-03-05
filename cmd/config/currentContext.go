package config

import (
	"github.com/deviceinsight/kafkactl/internal/global"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

func newCurrentContextCmd() *cobra.Command {
	var cmdCurrentContext = &cobra.Command{
		Use:     "current-context",
		Aliases: []string{"currentContext"},
		Short:   "show current context",
		Long:    `Displays the name of context that is currently active`,
		Run: func(_ *cobra.Command, _ []string) {
			context := global.GetCurrentContext()
			output.Infof("%s", context)
		},
	}

	return cmdCurrentContext
}
