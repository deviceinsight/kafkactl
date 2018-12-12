package cmd

import (
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
)

const VERSION = "0.0.1"

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "print the version of kafkactl",
	Run: func(cmd *cobra.Command, args []string) {
		output.Infof(VERSION)
	},
}

func init() {
	rootCmd.AddCommand(cmdVersion)
}
