package config

import (
	"github.com/deviceinsight/kafkactl/v5/internal/global"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/cobra"
)

func newCurrentContextCmd() *cobra.Command {
	var cmdCurrentContext = &cobra.Command{
		Use:     "current-context",
		Aliases: []string{"currentContext"},
		Short:   "show current context",
		Long:    `Displays the name of context that is currently active`,
		RunE: func(_ *cobra.Command, _ []string) error {
			context, err := global.GetCurrentContext()
			if err != nil {
				return err
			}
			output.Infof("%s", context)
			return nil
		},
	}

	return cmdCurrentContext
}
