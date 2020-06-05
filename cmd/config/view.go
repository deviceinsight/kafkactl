package config

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
)

func newViewCmd() *cobra.Command {

	var cmdView = &cobra.Command{
		Use:   "view",
		Short: "show contents of config file",
		Long:  `Shows the contents of the config file that is currently used`,
		Run: func(cmd *cobra.Command, args []string) {

			yamlFile, err := ioutil.ReadFile(viper.ConfigFileUsed())
			if err != nil {
				output.Fail(errors.Wrap(err, "unable to read config"))
			}

			output.Infof("%s", yamlFile)
		},
	}

	return cmdView
}
