package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectEditors(t *testing.T) {
	homeDir := t.TempDir()
	cwd := t.TempDir()

	claudeDir := filepath.Join(homeDir, ".config", "Claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(cwd, ".cursor"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cwd, ".cursor", "mcp.json"), []byte("{}"), 0o644))

	deps := initDeps{
		goos:       "linux",
		getwd:      func() (string, error) { return cwd, nil },
		homeDir:    func() (string, error) { return homeDir, nil },
		executable: func() (string, error) { return filepath.Join(cwd, "relay"), nil },
		lookPath: func(name string) (string, error) {
			if name == "code" {
				return filepath.Join(cwd, "code"), nil
			}
			return "", assert.AnError
		},
	}

	editors, err := detectEditors(deps)
	require.NoError(t, err)
	require.Len(t, editors, 3)

	assert.Equal(t, "Claude Desktop", editors[0].Name)
	assert.Equal(t, filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json"), editors[0].ConfigPath)
	assert.Equal(t, "Cursor", editors[1].Name)
	assert.Equal(t, filepath.Join(cwd, ".cursor", "mcp.json"), editors[1].ConfigPath)
	assert.Equal(t, "VS Code", editors[2].Name)
	assert.Equal(t, filepath.Join(cwd, ".vscode", "mcp.json"), editors[2].ConfigPath)
}

func TestConfigureEditor_MergesAndBacksUp(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ".cursor", "mcp.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, os.WriteFile(configPath, []byte(`{"mcpServers":{"other":{"command":"other"}}}`), 0o644))

	result, err := configureEditor(editorTarget{
		Name:       "Cursor",
		ConfigPath: configPath,
		RootKey:    "mcpServers",
	}, filepath.Join("C:", "Tools", "relay.exe"))
	require.NoError(t, err)

	assert.Equal(t, configPath, result.path)
	assert.Equal(t, configPath+".bak", result.backupPath)
	assert.False(t, result.alreadyConfigured)

	backup, err := os.ReadFile(configPath + ".bak")
	require.NoError(t, err)
	assert.JSONEq(t, `{"mcpServers":{"other":{"command":"other"}}}`, string(backup))

	var config map[string]map[string]map[string]string
	body, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &config))
	assert.Equal(t, "other", config["mcpServers"]["other"]["command"])
	assert.Equal(t, filepath.Join("C:", "Tools", "relay.exe"), config["mcpServers"]["relay"]["command"])
}

func TestConfigureEditor_AlreadyConfigured(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ".vscode", "mcp.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(configPath), 0o755))
	original := `{"servers":{"relay":{"command":"relay"}}}`
	require.NoError(t, os.WriteFile(configPath, []byte(original), 0o644))

	result, err := configureEditor(editorTarget{
		Name:       "VS Code",
		ConfigPath: configPath,
		RootKey:    "servers",
	}, filepath.Join("C:", "Tools", "relay.exe"))
	require.NoError(t, err)

	assert.True(t, result.alreadyConfigured)
	assert.Empty(t, result.backupPath)

	body, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.JSONEq(t, original, string(body))
}

func TestRunInitCommandWithDeps_AutoSelectsSingleEditor(t *testing.T) {
	homeDir := t.TempDir()
	cwd := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".config", "Claude"), 0o755))

	configPath := filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json")
	deps := initDeps{
		goos:       "linux",
		getwd:      func() (string, error) { return cwd, nil },
		homeDir:    func() (string, error) { return homeDir, nil },
		executable: func() (string, error) { return filepath.Join(cwd, "relay"), nil },
		lookPath:   func(string) (string, error) { return "", assert.AnError },
	}

	var stdout, stderr bytes.Buffer
	code := runInitCommandWithDeps(initOptions{}, bytes.NewBuffer(nil), &stdout, &stderr, deps)

	require.Equal(t, 0, code)
	assert.Empty(t, stderr.String())
	assert.Contains(t, stdout.String(), "relay init")
	assert.Contains(t, stdout.String(), "configuring Claude Desktop...")
	assert.Contains(t, stdout.String(), "done. restart Claude Desktop and relay is ready.")

	var config map[string]map[string]map[string]string
	body, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(body, &config))
	assert.Equal(t, filepath.Join(cwd, "relay"), config["mcpServers"]["relay"]["command"])
}
