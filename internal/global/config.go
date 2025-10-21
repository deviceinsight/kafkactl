package global

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
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

const configFileName = "config"
const writableConfigFileName = "current-context"

const ConfigEnvVariable = "KAFKA_CTL_CONFIG"
const WritableConfigEnvVariable = "KAFKA_CTL_WRITABLE_CONFIG"

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
	Init() error
	currentContext() string
	setCurrentContext(contextName string) error
	SetWritableConfig(viper *viper.Viper)
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

func GetCurrentContext() (string, error) {
	var context = configInstance.Flags().Context
	if context != "" {
		contexts := viper.GetStringMap("contexts")

		// check if it is an existing context
		if _, ok := contexts[context]; !ok {
			return "", fmt.Errorf("not a valid context: %s", context)
		}

		return context, nil
	}

	return configInstance.currentContext(), nil
}

func SetCurrentContext(contextName string) error {
	return configInstance.setCurrentContext(contextName)
}

func ResolvePath(filename string) (string, error) {

	wd, _ := os.Getwd()
	searchPaths := map[string]bool{wd: true}

	if _, err := os.Stat(filename); err == nil {
		return filename, err
	} else if !errors.Is(err, os.ErrNotExist) || filepath.IsAbs(filename) {
		return "", fmt.Errorf("unable to resolve path %q: %v", filename, err)
	}

	// relative to config file
	configDir := path.Dir(viper.ConfigFileUsed())
	if configDir != "." {
		searchPaths[configDir] = true
		cfgFilename := filepath.Join(path.Dir(viper.ConfigFileUsed()), filename)
		if _, err := os.Stat(cfgFilename); err == nil {
			return cfgFilename, err
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("unable to resolve path %q: %v", cfgFilename, err)
		}
	}

	// relative to EXTRA_PATHS
	if os.Getenv("EXTRA_PATHS") != "" {
		for _, extraPath := range strings.Split(os.Getenv("EXTRA_PATHS"), ":") {
			searchPaths[extraPath] = true
			cfgFilename := filepath.Join(extraPath, filename)
			if _, err := os.Stat(cfgFilename); err == nil {
				return cfgFilename, err
			} else if !errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("unable to resolve path %q: %v", cfgFilename, err)
			}
		}
	}

	// check in path (e.g. for kubectl binary)
	if _, err := exec.LookPath(filename); err == nil {
		return filename, nil
	}

	locations := ""
	for searchPath := range searchPaths {
		locations += searchPath + ","
	}

	return "", fmt.Errorf("cannot find %q in locations: [%s]", filename, strings.Trim(locations, ","))
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

func (c *config) SetWritableConfig(viper *viper.Viper) {
	c.writableConfig = viper
}

// Init reads in config file and ENV variables if set.
func (c *config) Init() error {

	viper.Reset()

	if c.flags.Verbose {
		output.IoStreams.EnableDebug()
	}

	configFile := resolveProjectConfigFileFromWorkingDir()

	switch {
	case c.flags.ConfigFile != "":
		configFile = &c.flags.ConfigFile
	case os.Getenv(ConfigEnvVariable) != "":
		envConfig := os.Getenv(ConfigEnvVariable)
		configFile = &envConfig
	}

	mapEnvVariables()

	if err := loadConfig(configFileName, viper.GetViper(), configFile); err != nil {
		if isUnknownError(err) {
			return fmt.Errorf("error reading config file: %s (%v)", viper.ConfigFileUsed(), err.Error())
		}
		err = generateDefaultConfig(configFileName, viper.GetViper())
		if err != nil {
			return fmt.Errorf("error generating default config file: %v", err.Error())
		}
	}

	var writableConfigFile *string

	if os.Getenv(WritableConfigEnvVariable) != "" {
		envConfig := os.Getenv(WritableConfigEnvVariable)
		writableConfigFile = &envConfig
	} else if os.Getenv(ConfigEnvVariable) != "" {
		// use config file provided via env as base path
		envConfig := filepath.Join(path.Dir(os.Getenv(ConfigEnvVariable)), writableConfigFileName+".yml")
		writableConfigFile = &envConfig
	}

	c.writableConfig = viper.New()
	if err := loadConfig(writableConfigFileName, c.writableConfig, writableConfigFile); err != nil {
		if isUnknownError(err) {
			return fmt.Errorf("error reading config file: %s (%v)", c.writableConfig.ConfigFileUsed(), err.Error())
		}
		err = generateDefaultConfig(writableConfigFileName, c.writableConfig)
		if err != nil {
			return fmt.Errorf("error generating default config file: %v", err.Error())
		}
	}
	return nil
}

func isUnknownError(err error) bool {

	var configFileNotFoundError viper.ConfigFileNotFoundError
	var pathError *os.PathError
	isConfigFileNotFoundError := errors.As(err, &configFileNotFoundError)
	isOsPathError := errors.As(err, &pathError)

	return !isConfigFileNotFoundError && !isOsPathError
}

func loadConfig(name string, viperInstance *viper.Viper, configFile *string) error {

	if configFile != nil {
		viperInstance.SetConfigFile(*configFile)
	} else {
		for _, p := range configPaths {
			viperInstance.AddConfigPath(os.ExpandEnv(p))
		}
		viperInstance.SetConfigName(name)
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

		p := workDir

		_, err = os.Stat(filepath.Join(p, projectConfigName))
		found := true

		for os.IsNotExist(err) {

			// stop when leaving a git repo
			if gitDir, statErr := os.Stat(filepath.Join(p, ".git")); statErr == nil && gitDir.IsDir() {
				found = false
				break
			}

			oldPath := p

			if p = filepath.Dir(oldPath); p == oldPath {
				output.Debugf("cannot find project config file: %s", projectConfigName)
				found = false
				break
			}
			_, err = os.Stat(filepath.Join(p, projectConfigName))
		}

		if found {
			configFile := filepath.Join(p, projectConfigName)
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
func generateDefaultConfig(name string, viperInstance *viper.Viper) error {

	filename := name + ".yml"
	cfgFile := filepath.Join(os.ExpandEnv(configPaths[0]), filename)

	if name == configFileName && os.Getenv(ConfigEnvVariable) != "" {
		// use config file provided via env
		cfgFile = os.Getenv(ConfigEnvVariable)
	} else if name == writableConfigFileName && os.Getenv(WritableConfigEnvVariable) != "" {
		// use config file provided via env
		cfgFile = os.Getenv(WritableConfigEnvVariable)
	} else if name == writableConfigFileName && os.Getenv(ConfigEnvVariable) != "" {
		// use config file provided via env as base path
		cfgFile = filepath.Join(path.Dir(os.Getenv(ConfigEnvVariable)), filename)
	} else if runtime.GOOS == "windows" {
		// use different configFile when running on windows
		for _, configPath := range configPaths {
			if strings.Contains(configPath, "$APPDATA") {
				cfgFile = filepath.Join(os.ExpandEnv(configPath), filename)
				break
			}
		}
	}

	if name == configFileName {
		viperInstance.SetDefault("contexts.default.brokers", []string{"localhost:9092"})
	} else {
		viperInstance.SetDefault("current-context", getInitialCurrentContext(cfgFile))
	}

	if err := os.MkdirAll(filepath.Dir(cfgFile), os.FileMode(0700)); err != nil {
		output.Warnf("cannot creating config file directory: %v", err)
		return nil
	}

	if err := viperInstance.WriteConfigAs(cfgFile); err != nil {
		output.Warnf("cannot write config file=%s: %v", cfgFile, err)
		return nil
	}

	output.Debugf("generated default config at %s", cfgFile)

	// We read generated config now
	if err := loadConfig(name, viperInstance, &cfgFile); err != nil {
		return fmt.Errorf("error reading config file: %s (%v)", viperInstance.ConfigFileUsed(), err.Error())
	}

	return nil
}

func getInitialCurrentContext(writableConfigFile string) string {

	if viper.GetString("current-context") != "" {
		output.Warnf(fmt.Sprintf(`the configuration of the current context is now managed in a separate file: %s
the parameter “current-context” in config.yml is no longer used. Please remove it to avoid confusion.`, writableConfigFile))
		return viper.GetString("current-context")
	}

	availableContexts := ListAvailableContexts()
	if len(availableContexts) > 0 {
		return availableContexts[0]
	}
	return ""
}
