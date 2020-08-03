package config

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

func newUseContextCmd() *cobra.Command {

	var cmdUseContext = &cobra.Command{
		Use:     "use-context",
		Aliases: []string{"useContext"},
		Short:   "switch active context",
		Long:    `command to switch active context`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			context := strings.Join(args, " ")

			contexts := viper.GetStringMap("contexts")

			// check if it is an existing context
			if _, ok := contexts[context]; !ok {
				output.Fail(errors.Errorf("not a valid context: %s", context))
			}

			viper.Set("current-context", context)

			if err := viper.WriteConfig(); err != nil {
				output.Fail(errors.Wrap(err, "unable to write config"))
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			contextMap := viper.GetStringMap("contexts")
			contexts := make([]string, 0, len(contextMap))
			for k := range contextMap {
				contexts = append(contexts, k)
			}

			return contexts, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmdUseContext
}
