package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/util"
	"github.com/deviceinsight/kafkactl/v5/pkg/plugins/auth"
)

var loadedPlugins = make(map[string]auth.AccessTokenProvider)

type pluginTokenProvider struct {
	pluginDelegate auth.AccessTokenProvider
}

func (p *pluginTokenProvider) Token() (*sarama.AccessToken, error) {
	token, err := p.pluginDelegate.Token()
	return &sarama.AccessToken{Token: token}, err
}

type scriptTokenProvider struct {
	script string
	args   []string
}

type scriptTokenResponse struct {
	Token      string            `json:"token"`
	Extensions map[string]string `json:"extensions,omitempty"`
}

func (p scriptTokenProvider) Token() (*sarama.AccessToken, error) {
	cmd := exec.Command(p.script, p.args...)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("token script %q failed with exit code %d: %s", p.script, exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute token script %q: %w", p.script, err)
	}

	var resp scriptTokenResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse token script output as JSON: %w (output was: %s)", err, string(output))
	}

	return &sarama.AccessToken{
		Token:      resp.Token,
		Extensions: resp.Extensions,
	}, nil
}

// newScriptTokenProvider create a token provider which uses a script to retrieve a token.
func newScriptTokenProvider(options map[string]any) (sarama.AccessTokenProvider, error) {
	scriptVal, ok := options["script"]
	if !ok {
		return nil, errors.New("missing required option 'script'")
	}
	script, ok := scriptVal.(string)
	if !ok {
		return nil, fmt.Errorf("option 'script' must be a string, got %T", scriptVal)
	}
	if script == "" {
		return nil, errors.New("option 'script' can't be empty")
	}

	var args []string
	if argsVal, ok := options["args"]; ok {
		args, ok = argsVal.([]string)
		if !ok {
			if argsAny, ok := argsVal.([]any); ok {
				args = make([]string, len(argsAny))
				for i, v := range argsAny {
					s, ok := v.(string)
					if !ok {
						return nil, fmt.Errorf("option 'args[%d]' must be a string, got %T", i, v)
					}
					args[i] = s
				}
			} else {
				return nil, fmt.Errorf("option 'args' must be a string array, got %T", argsVal)
			}
		}
	}

	return &scriptTokenProvider{
		script: script,
		args:   args,
	}, nil
}

func LoadTokenProviderPlugin(pluginName string, options map[string]any, brokers []string) (sarama.AccessTokenProvider, error) {
	if pluginName == "generic" {
		return newScriptTokenProvider(options)
	}

	loadedPlugin, ok := loadedPlugins[pluginName]
	if !ok {
		var err error
		loadedPlugin, err = util.LoadPlugin(pluginName, auth.TokenProviderPluginSpec)
		if err != nil {
			return nil, err
		}

		if options == nil {
			options = make(map[string]any)
		}

		if err := loadedPlugin.Init(options, brokers); err != nil {
			return nil, err
		}
		loadedPlugins[pluginName] = loadedPlugin
	}

	return &pluginTokenProvider{loadedPlugin}, nil
}
