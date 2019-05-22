package config

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
)

var cmdView = &cobra.Command{
	Use:   "view",
	Short: "show contents of config file",
	Long:  `Shows the contents of the config file that is currently used`,
	Run: func(cmd *cobra.Command, args []string) {

		yamlFile, err := ioutil.ReadFile(viper.ConfigFileUsed())
		if err != nil {
			output.Failf("unable to read config: %v", err)
		}

		output.Infof("%s", yamlFile)
	},
}

func init() {
}
