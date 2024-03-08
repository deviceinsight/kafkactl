package config

import (
	"github.com/deviceinsight/kafkactl/internal/global"
	"github.com/deviceinsight/kafkactl/internal/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newGetContextsCmd() *cobra.Command {

	var outputFormat string

	var cmdGetContexts = &cobra.Command{
		Use:     "get-contexts",
		Aliases: []string{"getContexts"},
		Short:   "list configured contexts",
		Long:    `Output names of all configured contexts`,
		Run: func(_ *cobra.Command, _ []string) {
			contexts := viper.GetStringMap("contexts")
			currentContext := global.GetCurrentContext()

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
