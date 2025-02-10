package global

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/spf13/viper"
)

type Flags struct {
	ConfigFile string
	Context    string
	Verbose    bool
}

const defaultContextPrefix = "CONTEXTS_DEFAULT_"
const GoContextKey = "global-config"

var projectConfigNames = []string{"kafkactl.yml", ".kafkactl.yml"}

var configPaths = []string{
	"$HOME/.config/kafkactl",
	"$HOME/.kafkactl",
	"$APPDATA/kafkactl",
	"/etc/kafkactl",
}

var configInstance *config

type Config interface {
	Flags() *Flags
	DefaultPaths() []string
	Init()
	currentContext() string
	setCurrentContext(contextName string) error
}

func NewConfig() Config {
	configInstance = &config{
		flags: Flags{},
	}
	return configInstance
}

func GetFlags() Flags {
	return configInstance.flags
}

func ListAvailableContexts() []string {

	var contexts []string
	for k := range viper.GetStringMap("contexts") {
		contexts = append(contexts, k)
	}

	sort.Strings(contexts)
	return contexts
}

func GetCurrentContext() string {
	var context = configInstance.Flags().Context
	if context != "" {
		contexts := viper.GetStringMap("contexts")

		// check if it is an existing context
		if _, ok := contexts[context]; !ok {
			output.Fail(fmt.Errorf("not a valid context: %s", context))
		}

		return context
	}

	return configInstance.currentContext()
}

func SetCurrentContext(contextName string) error {
	return configInstance.setCurrentContext(contextName)
}

type config struct {
	flags          Flags
	writableConfig *viper.Viper
}

func (c *config) Flags() *Flags {
	return &c.flags
}

func (c *config) DefaultPaths() []string {
	return configPaths
}

func (c *config) currentContext() string {
	return c.writableConfig.GetString("current-context")
}
func (c *config) setCurrentContext(contextName string) error {
	c.writableConfig.Set("current-context", contextName)
	return c.writableConfig.WriteConfig()
}

// Init reads in config file and ENV variables if set.
func (c *config) Init() {

	viper.Reset()

	if c.flags.Verbose {
		output.IoStreams.EnableDebug()
	}

	configFile := resolveProjectConfigFileFromWorkingDir()

	switch {
	case c.flags.ConfigFile != "":
		configFile = &c.flags.ConfigFile
	case os.Getenv("KAFKA_CTL_CONFIG") != "":
		envConfig := os.Getenv("KAFKA_CTL_CONFIG")
		configFile = &envConfig
	}

	mapEnvVariables()

	if err := c.loadConfig(viper.GetViper(), configFile); err != nil {
		if isUnknownError(err) {
			output.Failf("Error reading config file: %s (%v)", viper.ConfigFileUsed(), err.Error())
		}
		err = generateDefaultConfig()
		if err != nil {
			output.Failf("Error generating default config file: %v", err.Error())
		}

		// We read generated config now
		if err = c.loadConfig(viper.GetViper(), configFile); err != nil {
			output.Failf("Error reading config file: %s (%v)", viper.ConfigFileUsed(), err.Error())
		}
	}

	if configFile != nil && viper.GetString("current-context") == "" {
		// assuming the provided configFile is read-only
		c.writableConfig = viper.New()
		if err := c.loadConfig(c.writableConfig, nil); err != nil {
			if isUnknownError(err) {
				output.Failf("Error reading config file: %s (%v)", c.writableConfig.ConfigFileUsed(), err.Error())
			}
			err = generateDefaultConfig()
			if err != nil {
				output.Failf("Error generating default config file: %v", err.Error())
			}

			// We read generated config now
			if err = c.loadConfig(c.writableConfig, configFile); err != nil {
				output.Failf("Error reading config file: %s (%v)", viper.ConfigFileUsed(), err.Error())
			}
		}
	} else {
		c.writableConfig = viper.GetViper()
	}
}

func isUnknownError(err error) bool {

	var configFileNotFoundError viper.ConfigFileNotFoundError
	var pathError *os.PathError
	isConfigFileNotFoundError := errors.As(err, &configFileNotFoundError)
	isOsPathError := errors.As(err, &pathError)

	return !isConfigFileNotFoundError && !isOsPathError
}

func (c *config) loadConfig(viperInstance *viper.Viper, configFile *string) error {

	if configFile != nil {
		viperInstance.SetConfigFile(*configFile)
	} else {
		for _, path := range configPaths {
			viperInstance.AddConfigPath(os.ExpandEnv(path))
		}
		viperInstance.SetConfigName("config")
	}

	replacer := strings.NewReplacer("-", "_", ".", "_")
	viperInstance.SetEnvKeyReplacer(replacer)

	viperInstance.SetConfigType("yml")
	viperInstance.AutomaticEnv() // read in environment variables that match

	var err error
	if err = viperInstance.ReadInConfig(); err == nil {
		output.Debugf("Using config file: %s", viperInstance.ConfigFileUsed())
	}

	return err
}

func resolveProjectConfigFileFromWorkingDir() *string {

	workDir, err := os.Getwd()
	if err != nil {
		output.Debugf("cannot find project config file. unable to get working dir")
		return nil
	}

	for _, projectConfigName := range projectConfigNames {

		path := workDir

		_, err = os.Stat(filepath.Join(path, projectConfigName))
		found := true

		for os.IsNotExist(err) {

			// stop when leaving a git repo
			if gitDir, statErr := os.Stat(filepath.Join(path, ".git")); statErr == nil && gitDir.IsDir() {
				found = false
				break
			}

			oldPath := path

			if path = filepath.Dir(oldPath); path == oldPath {
				output.Debugf("cannot find project config file: %s", projectConfigName)
				found = false
				break
			}
			_, err = os.Stat(filepath.Join(path, projectConfigName))
		}

		if found {
			configFile := filepath.Join(path, projectConfigName)
			return &configFile
		}
	}

	return nil
}

func mapEnvVariables() {
	for _, short := range EnvVariables {
		long := defaultContextPrefix + short
		if os.Getenv(short) != "" && os.Getenv(long) == "" {
			_ = os.Setenv(long, os.Getenv(short))
		}
	}

	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, SaslTokenProviderOptions) {

			envKey := strings.SplitN(envVar, "=", 2)

			long := defaultContextPrefix + envKey[0]
			_ = os.Setenv(long, envKey[1])
		}
	}
}

// generateDefaultConfig generates default config in case there is no config
func generateDefaultConfig() error {

	cfgFile := filepath.Join(os.ExpandEnv(configPaths[0]), "config.yml")

	if os.Getenv("KAFKA_CTL_CONFIG") != "" {
		// use config file provided via env
		cfgFile = os.Getenv("KAFKA_CTL_CONFIG")
	} else if runtime.GOOS == "windows" {
		// use different configFile when running on windows
		for _, configPath := range configPaths {
			if strings.Contains(configPath, "$APPDATA") {
				cfgFile = filepath.Join(os.ExpandEnv(configPath), "config.yml")
				break
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(cfgFile), os.FileMode(0700)); err != nil {
		return err
	}

	viper.SetDefault("contexts.default.brokers", []string{"localhost:9092"})
	viper.SetDefault("current-context", "default")

	if err := viper.WriteConfigAs(cfgFile); err != nil {
		return err
	}

	output.Debugf("generated default config at %s", cfgFile)
	return nil
}
