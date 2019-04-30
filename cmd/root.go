package cmd

import (
	"fmt"
	"github.com/deviceinsight/kafkactl/cmd/alter"
	"github.com/deviceinsight/kafkactl/cmd/config"
	"github.com/deviceinsight/kafkactl/cmd/consume"
	"github.com/deviceinsight/kafkactl/cmd/create"
	"github.com/deviceinsight/kafkactl/cmd/deletion"
	"github.com/deviceinsight/kafkactl/cmd/describe"
	"github.com/deviceinsight/kafkactl/cmd/get"
	"github.com/deviceinsight/kafkactl/cmd/produce"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string
var Verbose bool

var rootCmd = &cobra.Command{
	Use:                    "kafkactl",
	BashCompletionFunction: bashCompletionFunc,
	Short:                  "command-line interface for Apache Kafka",
	Long:                   `A command-line interface the simplifies interaction with Kafka.`,
}

var configPaths = []string{"$HOME/.config/kafkactl", "$HOME/.kafkactl", "$SNAP_DATA/kafkactl", "/etc/kafkactl"}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(config.CmdConfig)
	rootCmd.AddCommand(consume.CmdConsume)
	rootCmd.AddCommand(create.CmdCreate)
	rootCmd.AddCommand(alter.CmdAlter)
	rootCmd.AddCommand(deletion.CmdDelete)
	rootCmd.AddCommand(describe.CmdDescribe)
	rootCmd.AddCommand(get.CmdGet)
	rootCmd.AddCommand(produce.CmdProduce)

	// use upper-case letters for shorthand params to avoid conflicts with local flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config-file", "C", "", fmt.Sprintf("config file. one of: %v", configPaths))
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "V", false, "verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else if os.Getenv("KAFKA_CTL_CONFIG") != "" {
		viper.SetConfigFile(os.Getenv("KAFKA_CTL_CONFIG"))
	} else {
		for _, path := range configPaths {
			viper.AddConfigPath(os.ExpandEnv(path))
		}
		viper.SetConfigName("config")
	}

	viper.SetConfigType("yml")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if Verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	} else {
		output.Failf("Error reading config file: %s (%v)", viper.ConfigFileUsed(), err)
	}
}
