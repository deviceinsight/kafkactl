package config

import (
	"github.com/deviceinsight/kafkactl/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var outputFormat string

var cmdGetContexts = &cobra.Command{
	Use:     "get-contexts",
	Aliases: []string{"getContexts"},
	Short:   "list configured contexts",
	Long:    `Output names of all configured contexts`,
	Run: func(cmd *cobra.Command, args []string) {
		contexts := viper.GetStringMap("contexts")
		currentContext := viper.GetString("current-context")

		if outputFormat == "compact" {
			for name := range contexts {
				output.Infof("%s", name)
			}
		} else {
			writer := output.CreateTableWriter()

			writer.WriteHeader("ACTIVE", "NAME")
			for context := range contexts {
				if currentContext == context {
					writer.Write("*", context)
				} else {
					writer.Write("", context)
				}
			}

			writer.Flush()
		}
	},
}

func init() {
	cmdGetContexts.Flags().StringVarP(&outputFormat, "output", "o", outputFormat, "Output format. One of: compact")
}
