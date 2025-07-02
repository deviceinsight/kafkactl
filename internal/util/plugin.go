package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/deviceinsight/kafkactl/v5/internal/global"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/deviceinsight/kafkactl/v5/pkg/plugins"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const PluginPathsEnvVariable = "KAFKA_CTL_PLUGIN_PATHS"

func LoadPlugin[P plugin.Plugin, I any](pluginName string, pluginSpec plugins.PluginSpec[P, I]) (impl I, err error) {

	pluginPath, err := resolvePluginPath(pluginName)
	if err != nil {
		return impl, err
	}

	output.Debugf("using plugin=%s location=%s", pluginName, pluginPath)

	pluginLogger := output.CreateVerboseLogger(pluginName, global.GetFlags().Verbose)

	logger := hclog.FromStandardLogger(pluginLogger, &hclog.LoggerOptions{
		Name:       "",
		Output:     os.Stderr,
		Level:      hclog.Debug,
		JSONFormat: false,
	})

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginSpec.Handshake,
		Plugins:         pluginSpec.GetMap(),
		Cmd:             exec.Command(pluginPath),
		Logger:          logger,
		Managed:         true,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return impl, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginSpec.InterfaceIdentifier)
	if err != nil {
		client.Kill()
		return impl, err
	}

	impl = raw.(I)

	return impl, nil
}

func resolvePluginPath(pluginName string) (string, error) {

	extension := ""

	if runtime.GOOS == "windows" {
		extension = ".exe"
	}

	pluginExecutable := fmt.Sprintf("kafkactl-%s-plugin%s", pluginName, extension)

	// search in paths configured via environment variable
	if os.Getenv(PluginPathsEnvVariable) != "" {
		pluginPaths := strings.Split(os.Getenv(PluginPathsEnvVariable), ":")
		for _, pluginPath := range pluginPaths {
			pluginLocation := filepath.Join(pluginPath, pluginExecutable)
			if _, err := os.Stat(pluginLocation); err == nil {
				return pluginLocation, nil
			} else if !os.IsNotExist(err) {
				return "", fmt.Errorf("error checking plugin path %q: %w", pluginPath, err)
			}
		}
	}

	// search relative to working dir
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	pluginPath := filepath.Join(workingDir, "../kafkactl-plugins/"+pluginName)
	pluginLocationWorkingDir := filepath.Join(pluginPath, pluginExecutable)

	if _, err = os.Stat(pluginLocationWorkingDir); err == nil {
		return pluginLocationWorkingDir, nil
	}

	// search in path
	if _, err := exec.LookPath(pluginExecutable); err == nil {
		return pluginExecutable, nil
	}

	// search next to executable
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	executablePath := filepath.Dir(executable)
	pluginLocationExe := filepath.Join(executablePath, pluginExecutable)

	if _, err := os.Stat(pluginLocationExe); err == nil {
		return pluginLocationExe, nil
	}

	return "", fmt.Errorf("plugin not found: %q", []string{
		pluginExecutable, pluginLocationExe, pluginLocationWorkingDir, os.Getenv(PluginPathsEnvVariable)})
}
