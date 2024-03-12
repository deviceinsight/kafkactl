package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/deviceinsight/kafkactl/v5/internal/global"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
	"github.com/deviceinsight/kafkactl/v5/pkg/plugins"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

func LoadPlugin[P plugin.Plugin, I any](pluginName string, pluginSpec plugins.PluginSpec[P, I]) (impl I, err error) {

	pluginPath, err := resolvePluginPath(pluginName)
	if err != nil {
		return impl, err
	}

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

	// search in path
	if _, err := os.Stat(pluginExecutable); err == nil {
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

	// search relative to working dir
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	pluginPath := filepath.Join(workingDir, "../kafkactl-plugins/"+pluginName)
	pluginLocationWorkingDir := filepath.Join(pluginPath, pluginExecutable)

	if _, err := os.Stat(pluginLocationWorkingDir); err == nil {
		return pluginLocationWorkingDir, nil
	}

	return "", errors.Wrapf(err, "plugin not found: %q", []string{pluginExecutable, pluginLocationExe, pluginLocationWorkingDir})
}
