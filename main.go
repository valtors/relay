package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"

	"github.com/valtors/relay/internal/logger"
	"github.com/valtors/relay/tools"
)

var Version = "dev"

var errHelpRequested = errors.New("help requested")

type startOptions struct {
	http bool
	addr string
}

type toolsOptions struct {
	json bool
}

type toolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	args, noColor := extractGlobalNoColorFlag(args)
	stdoutUI := newCLIUI(stdout, noColor)
	stderrUI := newCLIUI(stderr, noColor)

	if len(args) == 0 {
		return runStartCommand(nil, stdout, stderr, stderrUI, true)
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage(stdout, stdoutUI)
		return 0
	case "start":
		return runStartCommand(args[1:], stdout, stderr, stderrUI, false)
	case "tools":
		return runToolsCommand(args[1:], stdout, stderr, stdoutUI)
	case "init":
		return runInitCommand(args[1:], os.Stdin, stdout, stderr, stdoutUI)
	case "status":
		return runStatusCommand(args[1:], stdout, stderr, stdoutUI)
	case "doctor":
		return runDoctorCommand(args[1:], stdout, stderr, stdoutUI)
	case "version", "-v", "--version":
		return runVersionCommand(args[1:], stdout, stderr)
	default:
		if strings.HasPrefix(args[0], "-") {
			return runStartCommand(args, stdout, stderr, stderrUI, false)
		}

		fmt.Fprintf(stderr, "relay: '%s' is not a command. run 'relay help' if you're lost.\n\n", args[0])
		printUsage(stderr, stderrUI)
		return 1
	}
}

func runStartCommand(args []string, stdout, stderr io.Writer, ui cliUI, showBanner bool) int {
	opts, err := parseStartOptions(args)
	if err != nil {
		if errors.Is(err, errHelpRequested) {
			printStartUsage(stdout)
			return 0
		}

		fmt.Fprintf(stderr, "relay start: %v\n\n", err)
		printStartUsage(stderr)
		return 1
	}

	if showBanner && isTerminalWriter(stderr) {
		toolList := registeredTools()
		fmt.Fprint(stderr, ui.renderBanner(Version, len(toolList), "stdio"))
		fmt.Fprintf(stderr, "\n%s\n\n", ui.renderHint("stdio mode. run 'npx userelay tui' if you want the interactive ui."))
		logger.Quiet = true
	}

	if err := runStart(opts); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}

func runToolsCommand(args []string, stdout, stderr io.Writer, ui cliUI) int {
	opts, err := parseToolsOptions(args)
	if err != nil {
		if errors.Is(err, errHelpRequested) {
			printToolsUsage(stdout)
			return 0
		}

		fmt.Fprintf(stderr, "relay tools: %v\n\n", err)
		printToolsUsage(stderr)
		return 1
	}

	toolList := registeredTools()
	if opts.json {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(toolList); err != nil {
			fmt.Fprintf(stderr, "relay tools: %v\n", err)
			return 1
		}
		return 0
	}

	groups := groupToolsByCategory(toolList)
	categories := sortedCategoryNames(groups)

	fmt.Fprintf(stdout, "%s\n\n", ui.bold(fmt.Sprintf("relay tools (%d total)", len(toolList))))
	for i, category := range categories {
		entries := sortToolsForDisplay(groups[category])
		items := make([]string, 0, len(entries))
		for _, entry := range entries {
			items = append(items, shortToolName(entry))
		}
		icon, label := categoryMeta(category)
		fmt.Fprintf(stdout, "  %s %s (%d)\n", icon, ui.bold(label), len(entries))
		fmt.Fprintln(stdout, wrapJoinedItems(items, 76, "     "))
		if i < len(categories)-1 {
			fmt.Fprintln(stdout)
		}
	}

	return 0
}

func runStatusCommand(args []string, stdout, stderr io.Writer, ui cliUI) int {
	switch validateSimpleCommandArgs("status", args, stdout, stderr, printStatusUsage) {
	case commandArgsHelp:
		return 0
	case commandArgsInvalid:
		return 1
	}

	toolList := registeredTools()
	fmt.Fprint(stdout, ui.renderBanner(Version, len(toolList), "stdio | http"))
	return 0
}

func runVersionCommand(args []string, stdout, stderr io.Writer) int {
	switch validateSimpleCommandArgs("version", args, stdout, stderr, printVersionUsage) {
	case commandArgsHelp:
		return 0
	case commandArgsInvalid:
		return 1
	}

	fmt.Fprintf(stdout, "relay v%s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)
	return 0
}

func parseStartOptions(args []string) (startOptions, error) {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	opts := startOptions{}
	fs.BoolVar(&opts.http, "http", false, "serve over Streamable-HTTP instead of stdio")
	fs.StringVar(&opts.addr, "addr", ":8080", "HTTP listen address")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return startOptions{}, errHelpRequested
		}
		return startOptions{}, err
	}

	if fs.NArg() != 0 {
		return startOptions{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	return opts, nil
}

func parseToolsOptions(args []string) (toolsOptions, error) {
	fs := flag.NewFlagSet("tools", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	opts := toolsOptions{}
	fs.BoolVar(&opts.json, "json", false, "emit JSON output")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return toolsOptions{}, errHelpRequested
		}
		return toolsOptions{}, err
	}

	if fs.NArg() != 0 {
		return toolsOptions{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	return opts, nil
}

type commandArgsResult int

const (
	commandArgsOK commandArgsResult = iota
	commandArgsHelp
	commandArgsInvalid
)

func validateSimpleCommandArgs(name string, args []string, stdout, stderr io.Writer, usage func(io.Writer)) commandArgsResult {
	if len(args) == 0 {
		return commandArgsOK
	}

	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			usage(stdout)
			return commandArgsHelp
		}
	}

	fmt.Fprintf(stderr, "relay %s: unexpected arguments: %s\n\n", name, strings.Join(args, " "))
	usage(stderr)
	return commandArgsInvalid
}

func runStart(opts startOptions) error {
	_ = godotenv.Load()

	s := buildServer()
	if opts.http {
		return serveHTTP(s, opts.addr)
	}

	logger.Info("relay starting", "version", Version, "transport", "stdio")
	return server.ServeStdio(s)
}

func buildServer() *server.MCPServer {
	s := server.NewMCPServer(
		"relay",
		Version,
		server.WithToolCapabilities(true),
	)
	reg := tools.DefaultRegistry()
	reg.RegisterAll(s)
	logger.Info("tools loaded", "count", reg.Count())

	return s
}

func serveHTTP(s *server.MCPServer, addr string) error {
	httpSrv := server.NewStreamableHTTPServer(s)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("relay starting",
			"version", Version, "transport", "streamable-http",
			"addr", addr, "endpoint", addr+"/mcp",
		)
		if err := httpSrv.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpSrv.Shutdown(ctx)
	}
}

func registeredTools() []toolInfo {
	list := tools.DefaultRegistry().List()
	out := make([]toolInfo, 0, len(list))
	for _, entry := range list {
		out = append(out, toolInfo{
			Name:        entry.Definition.Name,
			Description: entry.Definition.Description,
			Category:    entry.Category,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Name < out[j].Name
	})

	return out
}

func groupToolsByCategory(toolList []toolInfo) map[string][]toolInfo {
	groups := make(map[string][]toolInfo, len(toolList))
	for _, entry := range toolList {
		groups[entry.Category] = append(groups[entry.Category], entry)
	}
	return groups
}

func sortedCategoryNames(groups map[string][]toolInfo) []string {
	names := make([]string, 0, len(groups))
	for category := range groups {
		names = append(names, category)
	}
	sort.Strings(names)
	return names
}

func categoryCount(toolList []toolInfo) int {
	seen := make(map[string]struct{}, len(toolList))
	for _, entry := range toolList {
		seen[entry.Category] = struct{}{}
	}
	return len(seen)
}

func truncateDescription(input string, max int) string {
	runes := []rune(strings.TrimSpace(input))
	if len(runes) <= max {
		return string(runes)
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

func printUsage(w io.Writer, ui cliUI) {
	fmt.Fprintln(w, ui.bold("Usage:"))
	fmt.Fprintln(w, "  relay [command] [flags]")
	fmt.Fprintln(w)

	printCommandGroup(w, ui, "Commands", []commandSummary{
		{usage: "relay", description: "Start the MCP server over stdio"},
		{usage: "relay start", description: "Start over stdio or Streamable HTTP"},
		{usage: "relay tools", description: "List available tools by category"},
		{usage: "relay init", description: "Detect and configure supported editors"},
		{usage: "relay config", description: "Print current configuration and environment"},
		{usage: "relay doctor", description: "Diagnose installation and config issues"},
		{usage: "relay status", description: "Show version, transports, and tool count"},
		{usage: "relay version", description: "Print relay version and platform"},
		{usage: "relay help", description: "Show this help message"},
	})

	printCommandGroup(w, ui, "Flags", []commandSummary{
		{usage: "--no-color", description: "Disable ANSI styling"},
	})

	printCommandGroup(w, ui, "Legacy start flags", []commandSummary{
		{usage: "-http", description: "Start the server in HTTP mode"},
		{usage: "-addr :9090", description: "Set the HTTP listen address"},
	})

	fmt.Fprintln(w, ui.bold("Examples"))
	fmt.Fprintln(w, "  relay")
	fmt.Fprintln(w, "  relay status")
	fmt.Fprintln(w, "  relay tools --json")
	fmt.Fprintln(w, "  relay init")
	fmt.Fprintln(w, "  relay config")
}

func printStartUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay")
	fmt.Fprintln(w, "  relay start [--http] [--addr :8080]")
	fmt.Fprintln(w, "  relay -http")
	fmt.Fprintln(w, "  relay -addr :9090")
}

func printToolsUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay tools [--json]")
}

func printInitUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay init [--list]")
}

func printStatusUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay status")
}

func printVersionUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay version")
}
