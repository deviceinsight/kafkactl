package config

import (
	"fmt"
	"os"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newViewCmd() *cobra.Command {

	var cmdView = &cobra.Command{
		Use:   "view",
		Short: "show contents of config file",
		Long:  `Shows the contents of the config file that is currently used`,
		RunE: func(_ *cobra.Command, _ []string) error {

			configFileUsed := viper.ConfigFileUsed()
			if _, err := os.Stat(configFileUsed); os.IsNotExist(err) {
				return fmt.Errorf("no config file loaded")
			}

			yamlFile, err := os.ReadFile(configFileUsed)
			if err != nil {
				return errors.Wrap(err, "unable to read config")
			}

			output.Infof("%s", yamlFile)
			return nil
		},
	}

	return cmdView
}
