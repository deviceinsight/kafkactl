package auth

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestScript(t *testing.T, content string, executable bool) string {
	t.Helper()
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-script.sh")

	perm := os.FileMode(0644)
	if executable {
		perm = 0755
	}

	err := os.WriteFile(scriptPath, []byte(content), perm)
	require.NoError(t, err)

	return scriptPath
}

func createTestProvider(script string, args []string) *scriptTokenProvider {
	return &scriptTokenProvider{
		logger: log.New(os.Stderr, "test", 0),
		script: script,
		args:   args,
	}
}

func TestScriptTokenProvider_Token_FileNotFound(t *testing.T) {
	provider := createTestProvider("/nonexistent/path/to/script.sh", []string{})

	token, err := provider.Token()

	assert.Nil(t, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestScriptTokenProvider_Token_FileNotExecutable(t *testing.T) {
	scriptContent := `#!/bin/bash
echo '{"token":"test-token"}'
`
	scriptPath := createTestScript(t, scriptContent, false)
	provider := createTestProvider(scriptPath, []string{})

	token, err := provider.Token()

	assert.Nil(t, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestScriptTokenProvider_Token_ErrorReturned(t *testing.T) {
	scriptContent := `#!/bin/bash
echo "Error message" >&2
exit 1
`
	scriptPath := createTestScript(t, scriptContent, true)
	provider := createTestProvider(scriptPath, []string{})

	token, err := provider.Token()

	assert.Nil(t, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token script")
	assert.Contains(t, err.Error(), "failed with exit code 1")
}

func TestScriptTokenProvider_Token_InvalidJSON(t *testing.T) {
	scriptContent := `#!/bin/bash
echo 'not valid json'
`
	scriptPath := createTestScript(t, scriptContent, true)
	provider := createTestProvider(scriptPath, []string{})

	token, err := provider.Token()

	assert.Nil(t, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse token script output as JSON")
	assert.Contains(t, err.Error(), "not valid json")
}

func TestScriptTokenProvider_Token_JSONWithoutTokenField(t *testing.T) {
	scriptContent := `#!/bin/bash
echo '{"extensions":{"key":"value"}}'
`
	scriptPath := createTestScript(t, scriptContent, true)
	provider := createTestProvider(scriptPath, []string{})

	token, err := provider.Token()

	assert.Nil(t, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token script returned empty token")
}

func TestScriptTokenProvider_Token_JSONWithEmptyToken(t *testing.T) {
	scriptContent := `#!/bin/bash
echo '{"token":""}'
`
	scriptPath := createTestScript(t, scriptContent, true)
	provider := createTestProvider(scriptPath, []string{})

	token, err := provider.Token()

	assert.Nil(t, token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token script returned empty token")
}

func TestScriptTokenProvider_Token_Success(t *testing.T) {
	scriptContent := `#!/bin/bash
echo '{"token":"my-access-token"}'
`
	scriptPath := createTestScript(t, scriptContent, true)
	provider := createTestProvider(scriptPath, []string{})

	token, err := provider.Token()

	assert.NoError(t, err)
	require.NotNil(t, token)
	assert.Equal(t, "my-access-token", token.Token)
	assert.Nil(t, token.Extensions)
}

func TestScriptTokenProvider_Token_SuccessWithExtensions(t *testing.T) {
	scriptContent := `#!/bin/bash
echo '{"token":"my-access-token","extensions":{"key1":"value1","key2":"value2"}}'
`
	scriptPath := createTestScript(t, scriptContent, true)
	provider := createTestProvider(scriptPath, []string{})

	token, err := provider.Token()

	assert.NoError(t, err)
	require.NotNil(t, token)
	assert.Equal(t, "my-access-token", token.Token)
	assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, token.Extensions)
}

func TestScriptTokenProvider_Token_WithArguments(t *testing.T) {
	scriptContent := `#!/bin/bash
echo '{"token":"token-with-args-'$1'-'$2'"}'
`
	scriptPath := createTestScript(t, scriptContent, true)
	provider := createTestProvider(scriptPath, []string{"arg1", "arg2"})

	token, err := provider.Token()

	assert.NoError(t, err)
	require.NotNil(t, token)
	assert.Equal(t, "token-with-args-arg1-arg2", token.Token)
}
