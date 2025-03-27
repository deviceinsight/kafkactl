package config

import (
	"github.com/deviceinsight/kafkactl/v5/internal/global"
	"github.com/pkg/errors"

	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newUseContextCmd() *cobra.Command {

	var cmdUseContext = &cobra.Command{
		Use:     "use-context",
		Aliases: []string{"useContext"},
		Short:   "switch active context",
		Long:    `command to switch active context`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {

			context := strings.Join(args, " ")

			contexts := viper.GetStringMap("contexts")

			// check if it is an existing context
			if _, ok := contexts[context]; !ok {
				return errors.Errorf("not a valid context: %s", context)
			}

			if err := global.SetCurrentContext(context); err != nil {
				return errors.Wrap(err, "unable to write config")
			}
			return nil
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return global.ListAvailableContexts(), cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmdUseContext
}
