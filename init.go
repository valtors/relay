package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type initOptions struct {
	list bool
}

type editorTarget struct {
	Name       string
	ConfigPath string
	RootKey    string
}

type initDeps struct {
	goos       string
	getwd      func() (string, error)
	homeDir    func() (string, error)
	executable func() (string, error)
	lookPath   func(string) (string, error)
}

var defaultInitDeps = initDeps{
	goos:       runtime.GOOS,
	getwd:      os.Getwd,
	homeDir:    os.UserHomeDir,
	executable: os.Executable,
	lookPath:   exec.LookPath,
}

func runInitCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts, err := parseInitOptions(args)
	if err != nil {
		if errors.Is(err, errHelpRequested) {
			printInitUsage(stdout)
			return 0
		}

		fmt.Fprintf(stderr, "relay init: %v\n\n", err)
		printInitUsage(stderr)
		return 1
	}

	return runInitCommandWithDeps(opts, stdin, stdout, stderr, defaultInitDeps)
}

func runInitCommandWithDeps(opts initOptions, stdin io.Reader, stdout, stderr io.Writer, deps initDeps) int {
	editors, err := detectEditors(deps)
	if err != nil {
		fmt.Fprintf(stderr, "relay init: %v\n", err)
		return 1
	}

	if opts.list {
		printInitHeader(stdout, true)
		printDetectedEditors(stdout, editors)
		return 0
	}

	fmt.Fprintln(stdout, "relay init")
	fmt.Fprintln(stdout)

	if len(editors) == 0 {
		fmt.Fprintln(stdout, "no supported editors detected.")
		return 1
	}

	printDetectedEditors(stdout, editors)
	fmt.Fprintln(stdout)

	selected, err := selectEditor(stdin, stdout, editors)
	if err != nil {
		fmt.Fprintf(stderr, "relay init: %v\n", err)
		return 1
	}

	exePath, err := deps.executable()
	if err != nil {
		fmt.Fprintf(stderr, "relay init: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "configuring %s...\n", selected.Name)
	result, err := configureEditor(*selected, filepath.Clean(exePath))
	if err != nil {
		fmt.Fprintf(stderr, "relay init: %v\n", err)
		return 1
	}

	if result.alreadyConfigured {
		fmt.Fprintf(stdout, "  already configured: %s\n\n", displayPath(result.path))
	} else {
		if result.backupPath != "" {
			fmt.Fprintf(stdout, "  backup: %s\n", displayPath(result.backupPath))
		}
		fmt.Fprintf(stdout, "  wrote: %s\n\n", displayPath(result.path))
	}

	fmt.Fprintf(stdout, "done. %s and relay is ready.\n", selected.NextStep())
	fmt.Fprintf(stdout, "%d tools available. try: \"resize screenshot.png to 800px wide\"\n", len(registeredTools()))
	return 0
}

func parseInitOptions(args []string) (initOptions, error) {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	opts := initOptions{}
	fs.BoolVar(&opts.list, "list", false, "list detected editors")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return initOptions{}, errHelpRequested
		}
		return initOptions{}, err
	}

	if fs.NArg() != 0 {
		return initOptions{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	return opts, nil
}

func detectEditors(deps initDeps) ([]editorTarget, error) {
	homeDir, err := deps.homeDir()
	if err != nil {
		return nil, err
	}

	cwd, err := deps.getwd()
	if err != nil {
		return nil, err
	}

	var editors []editorTarget

	if configPath := detectClaudeConfigPath(deps.goos, homeDir); configPath != "" && pathExists(configPath, true) {
		editors = append(editors, editorTarget{
			Name:       "Claude Desktop",
			ConfigPath: configPath,
			RootKey:    "mcpServers",
		})
	}

	if configPath := detectCursorConfigPath(homeDir, cwd, deps.lookPath); configPath != "" {
		editors = append(editors, editorTarget{
			Name:       "Cursor",
			ConfigPath: configPath,
			RootKey:    "mcpServers",
		})
	}

	if configPath := detectVSCodeConfigPath(cwd, deps.lookPath); configPath != "" {
		editors = append(editors, editorTarget{
			Name:       "VS Code",
			ConfigPath: configPath,
			RootKey:    "servers",
		})
	}

	return editors, nil
}

func detectClaudeConfigPath(goos, homeDir string) string {
	switch goos {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return filepath.Join(homeDir, "AppData", "Roaming", "Claude", "claude_desktop_config.json")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	default:
		return filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json")
	}
}

func detectCursorConfigPath(homeDir, cwd string, lookPath func(string) (string, error)) string {
	projectPath := filepath.Join(cwd, ".cursor", "mcp.json")
	if pathExists(projectPath, false) {
		return projectPath
	}

	homePath := filepath.Join(homeDir, ".cursor", "mcp.json")
	if pathExists(homePath, false) {
		return homePath
	}

	if commandExists(lookPath, "cursor", "cursor.exe") {
		return homePath
	}

	return ""
}

func detectVSCodeConfigPath(cwd string, lookPath func(string) (string, error)) string {
	projectPath := filepath.Join(cwd, ".vscode", "mcp.json")
	if pathExists(projectPath, false) {
		return projectPath
	}

	if commandExists(lookPath, "code", "code.cmd", "code.exe") {
		return projectPath
	}

	return ""
}

func commandExists(lookPath func(string) (string, error), names ...string) bool {
	for _, name := range names {
		if _, err := lookPath(name); err == nil {
			return true
		}
	}
	return false
}

func pathExists(path string, allowParentDir bool) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	if !allowParentDir {
		return false
	}

	_, err := os.Stat(filepath.Dir(path))
	return err == nil
}

func printInitHeader(w io.Writer, listMode bool) {
	if listMode {
		fmt.Fprintln(w, "relay init --list")
		fmt.Fprintln(w)
	}
}

func printDetectedEditors(w io.Writer, editors []editorTarget) {
	if len(editors) == 0 {
		fmt.Fprintln(w, "detected editors:")
		fmt.Fprintln(w, "  none")
		return
	}

	fmt.Fprintln(w, "detected editors:")
	for i, editor := range editors {
		fmt.Fprintf(w, "  %d. %s\n", i+1, editor.Name)
	}
}

func selectEditor(stdin io.Reader, stdout io.Writer, editors []editorTarget) (*editorTarget, error) {
	if len(editors) == 1 {
		return &editors[0], nil
	}

	scanner := bufio.NewScanner(stdin)
	for {
		fmt.Fprintf(stdout, "pick one (1-%d): ", len(editors))
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}

		choice := strings.TrimSpace(scanner.Text())
		index, err := strconv.Atoi(choice)
		if err != nil || index < 1 || index > len(editors) {
			fmt.Fprintln(stdout, "please enter a valid number.")
			continue
		}

		fmt.Fprintln(stdout)
		return &editors[index-1], nil
	}
}

type configureResult struct {
	path              string
	backupPath        string
	alreadyConfigured bool
}

func configureEditor(editor editorTarget, relayPath string) (configureResult, error) {
	existing, exists, err := loadConfig(editor.ConfigPath)
	if err != nil {
		return configureResult{}, err
	}

	root, err := ensureObject(existing, editor.RootKey)
	if err != nil {
		return configureResult{}, err
	}

	if _, ok := root["relay"]; ok {
		return configureResult{
			path:              editor.ConfigPath,
			alreadyConfigured: true,
		}, nil
	}

	root["relay"] = map[string]any{
		"command": relayPath,
	}
	existing[editor.RootKey] = root

	if err := os.MkdirAll(filepath.Dir(editor.ConfigPath), 0o755); err != nil {
		return configureResult{}, err
	}

	result := configureResult{path: editor.ConfigPath}
	if exists {
		backupPath := editor.ConfigPath + ".bak"
		if err := copyFile(editor.ConfigPath, backupPath); err != nil {
			return configureResult{}, err
		}
		result.backupPath = backupPath
	}

	body, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return configureResult{}, err
	}
	body = append(body, '\n')

	if err := os.WriteFile(editor.ConfigPath, body, 0o644); err != nil {
		return configureResult{}, err
	}

	return result, nil
}

func loadConfig(path string) (map[string]any, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, false, nil
		}
		return nil, false, err
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]any{}, true, nil
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, true, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}

	if config == nil {
		config = map[string]any{}
	}

	return config, true, nil
}

func ensureObject(config map[string]any, key string) (map[string]any, error) {
	value, ok := config[key]
	if !ok || value == nil {
		return map[string]any{}, nil
	}

	obj, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s in config must be a JSON object", key)
	}

	return obj, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func displayPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if rel, err := filepath.Rel(homeDir, path); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
		return filepath.Join("~", rel)
	}

	return path
}

func (e editorTarget) NextStep() string {
	switch e.Name {
	case "Claude Desktop":
		return "restart Claude Desktop"
	case "Cursor":
		return "restart Cursor"
	default:
		return "restart VS Code"
	}
}
