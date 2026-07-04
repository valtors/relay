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
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"

	"relay/internal/license"
	"relay/internal/logger"
	"relay/tools"
)

var Version = "dev"

var errHelpRequested = errors.New("help requested")

type apiKeyError struct{}

func (apiKeyError) Error() string {
	return "\n[relay] ANTHROPIC_API_KEY is not set.\nAdd it to your shell environment or a .env file in the project root."
}

type startupLicenseError struct {
	err error
}

func (e startupLicenseError) Error() string {
	return e.err.Error()
}

func (e startupLicenseError) Unwrap() error {
	return e.err
}

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
	if len(args) == 0 {
		return runStartCommand(nil, stdout, stderr)
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	case "start":
		return runStartCommand(args[1:], stdout, stderr)
	case "tools":
		return runToolsCommand(args[1:], stdout, stderr)
	case "status":
		return runStatusCommand(args[1:], stdout, stderr)
	case "version":
		return runVersionCommand(args[1:], stdout, stderr)
	default:
		if strings.HasPrefix(args[0], "-") {
			return runStartCommand(args, stdout, stderr)
		}

		fmt.Fprintf(stderr, "relay: unknown command %q\n\n", args[0])
		printUsage(stderr)
		return 1
	}
}

func runStartCommand(args []string, stdout, stderr io.Writer) int {
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

	if err := runStart(opts); err != nil {
		var missingAPIKey apiKeyError
		if errors.As(err, &missingAPIKey) {
			fmt.Fprintln(stderr, err)
			return 1
		}

		var licenseErr startupLicenseError
		if errors.As(err, &licenseErr) {
			fmt.Fprint(stderr, license.FriendlyMessage(licenseErr.err))
			return 1
		}

		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}

func runToolsCommand(args []string, stdout, stderr io.Writer) int {
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

	fmt.Fprintf(stdout, "relay tools (%d total)\n\n", len(toolList))
	for i, category := range categories {
		entries := groups[category]
		fmt.Fprintf(stdout, "%s (%d)\n", category, len(entries))
		for _, entry := range entries {
			fmt.Fprintf(stdout, "  %-18s %s\n", entry.Name, truncateDescription(entry.Description, 60))
		}
		if i < len(categories)-1 {
			fmt.Fprintln(stdout)
		}
	}

	return 0
}

func runStatusCommand(args []string, stdout, stderr io.Writer) int {
	switch validateSimpleCommandArgs("status", args, stdout, stderr, printStatusUsage) {
	case commandArgsHelp:
		return 0
	case commandArgsInvalid:
		return 1
	}

	toolList := registeredTools()
	fmt.Fprintf(stdout, "relay v%s\n", Version)
	fmt.Fprintf(stdout, "tools: %d registered (%d categories)\n", len(toolList), categoryCount(toolList))
	fmt.Fprintln(stdout, "transport: stdio (default) | http (with --http)")
	return 0
}

func runVersionCommand(args []string, stdout, stderr io.Writer) int {
	switch validateSimpleCommandArgs("version", args, stdout, stderr, printVersionUsage) {
	case commandArgsHelp:
		return 0
	case commandArgsInvalid:
		return 1
	}

	fmt.Fprintln(stdout, Version)
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

	lic, err := license.Verify()
	if err != nil {
		return startupLicenseError{err: err}
	}
	logger.Info("licensed", "subject", lic.Subject, "expires", lic.Expires)

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return apiKeyError{}
	}

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
		if err := httpSrv.Start(addr); err != nil && err != http.ErrServerClosed {
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

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay                 Start the MCP server in stdio mode")
	fmt.Fprintln(w, "  relay start [--http] [--addr :8080]")
	fmt.Fprintln(w, "  relay tools [--json]")
	fmt.Fprintln(w, "  relay status")
	fmt.Fprintln(w, "  relay version")
	fmt.Fprintln(w, "  relay help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Legacy flags:")
	fmt.Fprintln(w, "  relay -http           Start in HTTP mode")
	fmt.Fprintln(w, "  relay -addr :9090     Set HTTP address for legacy start mode")
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

func printStatusUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay status")
}

func printVersionUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay version")
}
