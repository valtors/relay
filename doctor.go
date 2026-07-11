package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type doctorOptions struct {
	fix bool
}

type doctorCheck struct {
	name    string
	status  string
	message string
}

func runDoctorCommand(args []string, stdout, stderr io.Writer, ui cliUI) int {
	opts, err := parseDoctorOptions(args)
	if err != nil {
		if errors.Is(err, errHelpRequested) {
			printDoctorUsage(stdout)
			return 0
		}
		fmt.Fprintf(stderr, "relay doctor: %v\n\n", err)
		printDoctorUsage(stderr)
		return 1
	}

	fmt.Fprintln(stdout, ui.bold("relay doctor"))
	fmt.Fprintln(stdout)

	checks := []doctorCheck{}
	checks = append(checks, checkBinaryHealth())
	checks = append(checks, checkNodeWrapper())
	checks = append(checks, checkEditorConfigs()...)
	checks = append(checks, checkEnvironment()...)
	checks = append(checks, checkNetwork())
	checks = append(checks, checkProcessLock())

	pass := 0
	warn := 0
	fail := 0
	for _, c := range checks {
		switch c.status {
		case "+":
			pass++
		case "!":
			warn++
		case "-":
			fail++
		}
		marker := ui.doctorMarker(c.status)
		fmt.Fprintf(stdout, "  %s %s\n", marker, c.name)
		fmt.Fprintf(stdout, "     %s\n", c.message)
	}

	fmt.Fprintln(stdout)
	if fail == 0 {
		fmt.Fprintf(stdout, "%s %s\n", ui.doctorMarker("+"), "all critical checks passed")
	} else {
		fmt.Fprintf(stdout, "%s %s\n", ui.doctorMarker("-"), fmt.Sprintf("found %d issue(s) to fix", fail))
	}

	if opts.fix {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, ui.renderHint("run: relay init to reconfigure your editor"))
		fmt.Fprintln(stdout, ui.renderHint("run: npm install -g userelay@latest if the binary is missing or out of date"))
	}

	if fail > 0 {
		return 1
	}
	return 0
}

func parseDoctorOptions(args []string) (doctorOptions, error) {
	if len(args) == 0 {
		return doctorOptions{}, nil
	}
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			return doctorOptions{}, errHelpRequested
		}
		if arg == "--fix" {
			return doctorOptions{fix: true}, nil
		}
	}
	return doctorOptions{}, fmt.Errorf("unknown arguments: %s", strings.Join(args, " "))
}

func printDoctorUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay doctor [--fix]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Diagnose common Relay installation and configuration issues.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --fix  Show suggested fixes after the report")
}

func (ui cliUI) doctorMarker(status string) string {
	if !ui.color {
		return status
	}
	switch status {
	case "+":
		return ui.green("+")
	case "!":
		if r := ui.lipglossRenderer(); r != nil {
			return r.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render("!")
		}
		return "!"
	case "-":
		if r := ui.lipglossRenderer(); r != nil {
			return r.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("-")
		}
		return "-"
	default:
		return status
	}
}

func checkBinaryHealth() doctorCheck {
	exePath, err := os.Executable()
	if err != nil {
		return doctorCheck{"relay binary", "-", fmt.Sprintf("cannot locate relay binary: %v", err)}
	}

	info, err := os.Stat(exePath)
	if err != nil {
		return doctorCheck{"relay binary", "-", fmt.Sprintf("cannot stat binary at %s: %v", displayPath(exePath), err)}
	}

	cmd := exec.Command(exePath, "version")
	out, err := cmd.Output()
	versionOutput := strings.TrimSpace(string(out))
	if err != nil {
		return doctorCheck{"relay binary", "-", fmt.Sprintf("binary at %s exists but cannot run: %v", displayPath(exePath), err)}
	}

	_ = info
	return doctorCheck{"relay binary", "+", fmt.Sprintf("%s at %s", versionOutput, displayPath(exePath))}
}

func checkNodeWrapper() doctorCheck {
	npmRelay, err := exec.LookPath("relay")
	if err != nil {
		return doctorCheck{"npm wrapper", "!", "npm wrapper script not found in PATH; run `npm install -g userelay` to install"}
	}

	cmd := exec.Command(npmRelay, "version")
	out, err := cmd.Output()
	if err != nil {
		return doctorCheck{"npm wrapper", "-", fmt.Sprintf("npm wrapper found at %s but failed: %v", displayPath(npmRelay), err)}
	}

	wrapperVersion := strings.TrimSpace(string(out))
	goVersion := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	_ = goVersion

	if wrapperVersion != "" && !strings.Contains(wrapperVersion, "dev") {
		return doctorCheck{"npm wrapper", "+", fmt.Sprintf("%s at %s", wrapperVersion, displayPath(npmRelay))}
	}

	return doctorCheck{"npm wrapper", "+", fmt.Sprintf("installed at %s", displayPath(npmRelay))}
}

func checkEditorConfigs() []doctorCheck {
	var checks []doctorCheck
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []doctorCheck{{"editor configs", "-", fmt.Sprintf("cannot determine home directory: %v", err)}}
	}

	editors := []struct {
		name     string
		path     func(string) string
		rootKey  string
	}{
		{
			name: "Claude Desktop",
			path: func(h string) string { return detectClaudeConfigPath(runtime.GOOS, h) },
			rootKey: "mcpServers",
		},
		{
			name: "Cursor",
			path: func(h string) string {
				cwd, _ := os.Getwd()
				return detectCursorConfigPath(h, cwd, exec.LookPath)
			},
			rootKey: "mcpServers",
		},
		{
			name: "VS Code",
			path: func(h string) string {
				cwd, _ := os.Getwd()
				return detectVSCodeConfigPath(cwd, exec.LookPath)
			},
			rootKey: "servers",
		},
	}

	for _, editor := range editors {
		cfgPath := editor.path(homeDir)
		if cfgPath == "" {
			checks = append(checks, doctorCheck{editor.name, "!", "not detected"})
			continue
		}

		data, err := os.ReadFile(cfgPath)
		if err != nil {
			if os.IsNotExist(err) {
				checks = append(checks, doctorCheck{editor.name, "!", fmt.Sprintf("config not found at %s", displayPath(cfgPath))})
			} else {
				checks = append(checks, doctorCheck{editor.name, "-", fmt.Sprintf("cannot read config at %s: %v", displayPath(cfgPath), err)})
			}
			continue
		}

		var cfg map[string]any
		if err := json.Unmarshal(data, &cfg); err != nil {
			checks = append(checks, doctorCheck{editor.name, "-", fmt.Sprintf("invalid JSON at %s: %v", displayPath(cfgPath), err)})
			continue
		}

		root, ok := cfg[editor.rootKey].(map[string]any)
		if !ok {
			checks = append(checks, doctorCheck{editor.name, "!", fmt.Sprintf("config exists but has no %s entry", editor.rootKey)})
			continue
		}

		if _, ok := root["relay"]; !ok {
			checks = append(checks, doctorCheck{editor.name, "!", fmt.Sprintf("relay not configured in %s", displayPath(cfgPath))})
			continue
		}

		relayCfg, ok := root["relay"].(map[string]any)
		if !ok {
			checks = append(checks, doctorCheck{editor.name, "-", fmt.Sprintf("relay entry is not an object in %s", displayPath(cfgPath))})
			continue
		}

		cmd, ok := relayCfg["command"].(string)
		if !ok || cmd == "" {
			checks = append(checks, doctorCheck{editor.name, "-", fmt.Sprintf("relay entry missing command in %s", displayPath(cfgPath))})
			continue
		}

		if _, err := exec.LookPath(cmd); err != nil && !filepath.IsAbs(cmd) {
			checks = append(checks, doctorCheck{editor.name, "!", fmt.Sprintf("relay command %s not found in PATH", cmd)})
			continue
		}

		if !pathExists(cmd, false) && !commandExists(exec.LookPath, cmd) {
			checks = append(checks, doctorCheck{editor.name, "!", fmt.Sprintf("relay command %s does not exist", cmd)})
			continue
		}

		checks = append(checks, doctorCheck{editor.name, "+", fmt.Sprintf("configured at %s", displayPath(cfgPath))})
	}

	return checks
}

func checkEnvironment() []doctorCheck {
	var checks []doctorCheck

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		checks = append(checks, doctorCheck{"ANTHROPIC_API_KEY", "+", "set (workflow tools enabled)"})
	} else {
		checks = append(checks, doctorCheck{"ANTHROPIC_API_KEY", "!", "not set; workflow tools require this key"})
	}

	if os.Getenv("GITHUB_TOKEN") != "" {
		checks = append(checks, doctorCheck{"GITHUB_TOKEN", "+", "set (helps npm downloads and workflow tools)"})
	} else {
		checks = append(checks, doctorCheck{"GITHUB_TOKEN", "!", "not set; optional, but helps avoid rate limits"})
	}

	return checks
}

func checkNetwork() doctorCheck {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/valtors/relay/releases/latest")
	if err != nil {
		return doctorCheck{"GitHub connectivity", "!", fmt.Sprintf("cannot reach GitHub releases: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return doctorCheck{"GitHub connectivity", "!", fmt.Sprintf("GitHub returned status %d", resp.StatusCode)}
	}
	return doctorCheck{"GitHub connectivity", "+", "can reach GitHub releases"}
}

func checkProcessLock() doctorCheck {
	if runtime.GOOS != "windows" {
		return doctorCheck{"process lock", "+", "not applicable on this platform"}
	}

	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq relay.exe", "/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return doctorCheck{"process lock", "!", fmt.Sprintf("could not check running processes: %v", err)}
	}

	output := string(out)
	if strings.Contains(output, "relay.exe") {
		return doctorCheck{"process lock", "!", "relay.exe is currently running; installs on Windows may fail with EPERM until you stop it"}
	}

	return doctorCheck{"process lock", "+", "no running relay.exe process found"}
}
