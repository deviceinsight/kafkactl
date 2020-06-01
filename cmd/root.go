package cmd

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/deviceinsight/kafkactl/cmd/alter"
	"github.com/deviceinsight/kafkactl/cmd/config"
	"github.com/deviceinsight/kafkactl/cmd/consume"
	"github.com/deviceinsight/kafkactl/cmd/create"
	"github.com/deviceinsight/kafkactl/cmd/deletion"
	"github.com/deviceinsight/kafkactl/cmd/describe"
	"github.com/deviceinsight/kafkactl/cmd/get"
	"github.com/deviceinsight/kafkactl/cmd/produce"
	"github.com/deviceinsight/kafkactl/cmd/reset"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func KafkactlCommand(streams output.IOStreams) *cobra.Command {
	output.IoStreams = streams
	rootCmd.SetOut(streams.Out)
	rootCmd.SetErr(streams.ErrOut)
	return rootCmd
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
	rootCmd.AddCommand(reset.CmdReset)

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

	if Verbose {
		sarama.Logger = log.New(os.Stderr, "[sarama  ] ", log.LstdFlags)
		output.DebugLogger = log.New(os.Stderr, "[kafkactl] ", log.LstdFlags)
	}

	viper.SetConfigType("yml")
	viper.AutomaticEnv() // read in environment variables that match

	if err := readConfig(); err != nil {
		output.Fail(err)
	}
}

func readConfig() error {
	var err error
	if err = viper.ReadInConfig(); err == nil {
		output.Debugf("Using config file: %s", viper.ConfigFileUsed())
		return nil
	}

	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		return errors.Errorf("Error reading config file: %s (%v)", viper.ConfigFileUsed(), err)
	} else {
		err = generateDefaultConfig()
		if err != nil {
			return errors.Wrap(err, "Error generating default config: ")
		}
	}

	// We read generated config now
	if err = viper.ReadInConfig(); err == nil {
		output.Debugf("Using config file: %s", viper.ConfigFileUsed())
		return nil
	} else {
		return errors.Errorf("Error reading config file: %s (%v)", viper.ConfigFileUsed(), err)
	}
}

// generateDefaultConfig generates default config in case there is no config
func generateDefaultConfig() error {
	if err := os.MkdirAll(os.ExpandEnv(configPaths[0]), os.FileMode(0700)); err != nil {
		return err
	}
	pathToConfig := filepath.Join(os.ExpandEnv(configPaths[0]), "config.yml")
	f, err := os.Create(pathToConfig)
	if err != nil {
		return fmt.Errorf("failed to generate default config at %s", pathToConfig)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	defaultConfigContent := `
contexts:
  localhost:
    brokers:
    - localhost:9092
current-context: localhost`

	if os.Getenv("BROKER") != "" {
		// this is useful for running in docker
		defaultConfigContent = strings.Replace(defaultConfigContent, "localhost:9092", os.Getenv("BROKER"), -1)
	}

	_, err = f.WriteString(defaultConfigContent)

	if err == nil {
		output.Debugf("generated default config at %s", pathToConfig)
	}

	return err
}
