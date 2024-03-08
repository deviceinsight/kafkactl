package config

import (
	"sort"

	"github.com/deviceinsight/kafkactl/v5/internal/global"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
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
		Run: func(_ *cobra.Command, args []string) {

			context := strings.Join(args, " ")

			contexts := viper.GetStringMap("contexts")

			// check if it is an existing context
			if _, ok := contexts[context]; !ok {
				output.Fail(errors.Errorf("not a valid context: %s", context))
			}

			if err := global.SetCurrentContext(context); err != nil {
				output.Fail(errors.Wrap(err, "unable to write config"))
			}
		},
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			contextMap := viper.GetStringMap("contexts")
			contexts := make([]string, 0, len(contextMap))
			for k := range contextMap {
				contexts = append(contexts, k)
			}

			sort.Strings(contexts)

			return contexts, cobra.ShellCompDirectiveNoFileComp
		},
	}

	return cmdUseContext
}
