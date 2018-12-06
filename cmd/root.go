// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/random-dwi/kafkactl/cmd/config"
	"github.com/random-dwi/kafkactl/cmd/consume"
	"github.com/random-dwi/kafkactl/cmd/create"
	"github.com/random-dwi/kafkactl/cmd/deletion"
	"github.com/random-dwi/kafkactl/cmd/get"
	"github.com/random-dwi/kafkactl/cmd/produce"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string
var Verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "kafkactl",
	BashCompletionFunction: bashCompletionFunc,
	Short: "command-line interface for Apache Kafka",
	Long:  `A command-line interface the simplifies interaction with Kafka.`,
}

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

	rootCmd.AddCommand(get.GetCmd)
	rootCmd.AddCommand(deletion.DeleteCmd)
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(consume.ConsumeCmd)
	rootCmd.AddCommand(produce.ProduceCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kafkactl.yaml)")

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kafkactl" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kafkactl")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if Verbose {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	} else {
		if Verbose {
			fmt.Println("Error reading config file:", viper.ConfigFileUsed(), err)
		}
	}
}
