package config

import (
	"github.com/deviceinsight/kafkactl/v5/internal/global"
	"github.com/deviceinsight/kafkactl/v5/internal/output"

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
		RunE: func(_ *cobra.Command, _ []string) error {
			contexts := viper.GetStringMap("contexts")
			currentContext, err := global.GetCurrentContext()
			if err != nil {
				return err
			}
			if outputFormat == "compact" {
				for name := range contexts {
					output.Infof("%s", name)
				}
			} else {
				writer := output.CreateTableWriter()

				if err := writer.WriteHeader("ACTIVE", "NAME"); err != nil {
					return err
				}
				for context := range contexts {
					if currentContext == context {
						if err := writer.Write("*", context); err != nil {
							return err
						}
					} else {
						if err := writer.Write("", context); err != nil {
							return err
						}
					}
				}

				if err := writer.Flush(); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmdGetContexts.Flags().StringVarP(&outputFormat, "output", "o", outputFormat, "output format. One of: compact")

	return cmdGetContexts
}
