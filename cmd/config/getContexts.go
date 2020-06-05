package config

import (
	"github.com/deviceinsight/kafkactl/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var outputFormat string

func newGetContextsCmd() *cobra.Command {

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

				if err := writer.WriteHeader("ACTIVE", "NAME"); err != nil {
					output.Fail(err)
				}
				for context := range contexts {
					if currentContext == context {
						if err := writer.Write("*", context); err != nil {
							output.Fail(err)
						}
					} else {
						if err := writer.Write("", context); err != nil {
							output.Fail(err)
						}
					}
				}

				if err := writer.Flush(); err != nil {
					output.Fail(err)
				}
			}
		},
	}

	cmdGetContexts.Flags().StringVarP(&outputFormat, "output", "o", outputFormat, "output format. One of: compact")

	return cmdGetContexts
}
