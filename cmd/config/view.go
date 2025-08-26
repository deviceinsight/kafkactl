package config

import (
	"bytes"
	"fmt"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newViewCmd() *cobra.Command {

	var cmdView = &cobra.Command{
		Use:   "view",
		Short: "show current config",
		Long:  `Shows the merged config that is currently used`,
		RunE: func(_ *cobra.Command, _ []string) error {

			c := new(bytes.Buffer)

			if err := viper.WriteConfigTo(c); err != nil {
				return fmt.Errorf("unable to write config: %v", err)
			}

			output.Infof("%s", c.String())
			return nil
		},
	}

	return cmdView
}
