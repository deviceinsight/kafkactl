package cmd

import (
	"fmt"
	"github.com/deviceinsight/kafkactl/cmd/alter"
	"github.com/deviceinsight/kafkactl/cmd/attach"
	"github.com/deviceinsight/kafkactl/cmd/config"
	"github.com/deviceinsight/kafkactl/cmd/consume"
	"github.com/deviceinsight/kafkactl/cmd/create"
	"github.com/deviceinsight/kafkactl/cmd/deletion"
	"github.com/deviceinsight/kafkactl/cmd/describe"
	"github.com/deviceinsight/kafkactl/cmd/get"
	"github.com/deviceinsight/kafkactl/cmd/produce"
	"github.com/deviceinsight/kafkactl/cmd/reset"
	"github.com/deviceinsight/kafkactl/operations/k8s"
	"github.com/deviceinsight/kafkactl/output"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var cfgFile string
var Verbose bool

var envMapping = map[string]string{
	"REQUESTTIMEOUT":      "CONTEXTS_DEFAULT_REQUESTTIMEOUT",
	"BROKERS":             "CONTEXTS_DEFAULT_BROKERS",
	"TLS_ENABLED":         "CONTEXTS_DEFAULT_TLS_ENABLED",
	"TLS_CA":              "CONTEXTS_DEFAULT_TLS_CA",
	"TLS_CERT":            "CONTEXTS_DEFAULT_TLS_CERT",
	"TLS_CERTKEY":         "CONTEXTS_DEFAULT_TLS_CERTKEY",
	"TLS_INSECURE":        "CONTEXTS_DEFAULT_TLS_INSECURE",
	"SASL_ENABLED":        "CONTEXTS_DEFAULT_SASL_ENABLED",
	"SASL_USERNAME":       "CONTEXTS_DEFAULT_SASL_USERNAME",
	"SASL_PASSWORD":       "CONTEXTS_DEFAULT_SASL_PASSWORD",
	"SASL_MECHANISM":      "CONTEXTS_DEFAULT_SASL_MECHANISM",
	"CLIENTID":            "CONTEXTS_DEFAULT_CLIENTID",
	"KAFKAVERSION":        "CONTEXTS_DEFAULT_KAFKAVERSION",
	"AVRO_SCHEMAREGISTRY": "CONTEXTS_DEFAULT_AVRO_SCHEMAREGISTRY",
	"DEFAULTPARTITIONER":  "CONTEXTS_DEFAULT_DEFAULTPARTITIONER",
}

var configPaths = []string{"$HOME/.config/kafkactl", "$HOME/.kafkactl", "$SNAP_REAL_HOME/.config/kafkactl", "$SNAP_DATA/kafkactl", "/etc/kafkactl"}

func NewKafkactlCommand(streams output.IOStreams) *cobra.Command {

	var rootCmd = &cobra.Command{
		Use:   "kafkactl",
		Short: "command-line interface for Apache Kafka",
		Long:  `A command-line interface the simplifies interaction with Kafka.`,
	}

	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(config.NewConfigCmd())
	rootCmd.AddCommand(consume.NewConsumeCmd())
	rootCmd.AddCommand(create.NewCreateCmd())
	rootCmd.AddCommand(alter.NewAlterCmd())
	rootCmd.AddCommand(deletion.NewDeleteCmd())
	rootCmd.AddCommand(describe.NewDescribeCmd())
	rootCmd.AddCommand(get.NewGetCmd())
	rootCmd.AddCommand(produce.NewProduceCmd())
	rootCmd.AddCommand(reset.NewResetCmd())
	rootCmd.AddCommand(attach.NewAttachCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newDocsCmd())

	// use upper-case letters for shorthand params to avoid conflicts with local flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config-file", "C", "", fmt.Sprintf("config file. one of: %v", configPaths))
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "V", false, "verbose output")

	k8s.KafkaCtlVersion = Version

	output.IoStreams = streams
	rootCmd.SetOut(streams.Out)
	rootCmd.SetErr(streams.ErrOut)
	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.Reset()

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
		output.IoStreams.EnableDebug()
	}

	if Verbose && os.Getenv("SNAP_NAME") != "" {
		output.Debugf("Running snap version %s on %s", os.Getenv("SNAP_VERSION"), os.Getenv("SNAP_ARCH"))
	}

	mapEnvVariables()

	replacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetDefault("contexts.default.brokers", []string{"localhost:9092"})
	viper.SetDefault("current-context", "default")

	viper.SetConfigType("yml")
	viper.AutomaticEnv() // read in environment variables that match

	if err := readConfig(); err != nil {
		output.Fail(err)
	}
}

func mapEnvVariables() {
	for short, long := range envMapping {
		if os.Getenv(short) != "" && os.Getenv(long) == "" {
			_ = os.Setenv(long, os.Getenv(short))
		}
	}
}

func readConfig() error {
	var err error
	if err = viper.ReadInConfig(); err == nil {
		output.Debugf("Using config file: %s", viper.ConfigFileUsed())
		return nil
	}

	_, isConfigFileNotFoundError := err.(viper.ConfigFileNotFoundError)
	_, isOsPathError := err.(*os.PathError)

	if !isConfigFileNotFoundError && !isOsPathError {
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

	cfgFile := filepath.Join(os.ExpandEnv(configPaths[0]), "config.yml")

	if os.Getenv("KAFKA_CTL_CONFIG") != "" {
		// use config file provided via env
		cfgFile = os.Getenv("KAFKA_CTL_CONFIG")
	} else if os.Getenv("SNAP_REAL_HOME") != "" {
		// use different configFile when running in snap
		for _, configPath := range configPaths {
			if strings.Contains(configPath, "$SNAP_REAL_HOME") {
				cfgFile = filepath.Join(os.ExpandEnv(configPath), "config.yml")
				break
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(cfgFile), os.FileMode(0700)); err != nil {
		return err
	}

	if err := viper.WriteConfigAs(cfgFile); err != nil {
		return err
	}

	output.Debugf("generated default config at %s", cfgFile)
	return nil
}
