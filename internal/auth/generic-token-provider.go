package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/IBM/sarama"
	"github.com/deviceinsight/kafkactl/v5/internal/global"
	"github.com/deviceinsight/kafkactl/v5/internal/output"
)

type scriptTokenProvider struct {
	logger *log.Logger
	script string
	args   []string
}

type scriptTokenResponse struct {
	Token      string            `json:"token"`
	Extensions map[string]string `json:"extensions,omitempty"`
}

func (p *scriptTokenProvider) Token() (*sarama.AccessToken, error) {
	cmd := exec.Command(p.script, p.args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()

	if stderr.Len() > 0 {
		scanner := bufio.NewScanner(&stderr)
		for scanner.Scan() {
			if err != nil || global.GetFlags().Verbose {
				p.logger.Println(scanner.Text())
			}
		}
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("token script %q failed with exit code %d", p.script, exitErr.ExitCode())
		}
		return nil, fmt.Errorf("failed to execute token script %q: %w", p.script, err)
	}

	var resp scriptTokenResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse token script output as JSON: %w (output was: %s)", err, string(out))
	}

	if resp.Token == "" {
		return nil, fmt.Errorf("token script returned empty token (output was: %s)", string(out))
	}

	return &sarama.AccessToken{
		Token:      resp.Token,
		Extensions: resp.Extensions,
	}, nil
}

// newGenericTokenProvider create a token provider which uses a script to retrieve a token.
func newGenericTokenProvider(options map[string]any) (sarama.AccessTokenProvider, error) {

	output.Debugf("using plugin=generic")
	logger := output.CreateLogger(os.Stderr, "script")

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

	scriptPath, err := global.ResolvePath(script)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path %q: %w", script, err)
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
		logger: logger,
		script: scriptPath,
		args:   args,
	}, nil
}
