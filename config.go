package main

import (
	"fmt"
	"io"
	"os"
)

func runConfigCommand(args []string, stdout, stderr io.Writer, ui cliUI) int {
	switch validateSimpleCommandArgs("config", args, stdout, stderr, printConfigUsage) {
	case commandArgsHelp:
		return 0
	case commandArgsInvalid:
		return 1
	}

	toolList := registeredTools()

	fmt.Fprintf(stdout, "%s\n", ui.bold("relay config"))
	fmt.Fprintf(stdout, "  version:     %s\n", Version)
	fmt.Fprintf(stdout, "  transport:   stdio (default) | http (with --http)\n")
	fmt.Fprintf(stdout, "  tools:       %d registered (%d categories)\n\n", len(toolList), categoryCount(toolList))

	fmt.Fprintf(stdout, "%s\n", ui.bold("environment"))

	envKeys := []struct {
		key, label string
	}{
		{"ANTHROPIC_API_KEY", "ANTHROPIC_API_KEY"},
		{"RELAY_WEB_TIMEOUT", "RELAY_WEB_TIMEOUT"},
		{"RELAY_HTTP_ADDR", "RELAY_HTTP_ADDR"},
		{"CHECKPOINT_TIMEOUT_MINUTES", "CHECKPOINT_TIMEOUT_MINUTES"},
		{"MAX_ITERATIONS_PER_STAGE", "MAX_ITERATIONS_PER_STAGE"},
	}

	for _, env := range envKeys {
		val := os.Getenv(env.key)
		if val == "" {
			fmt.Fprintf(stdout, "  %-28s not set\n", env.label)
		} else if env.key == "ANTHROPIC_API_KEY" {
			fmt.Fprintf(stdout, "  %-28s %s...\n", env.label, val[:min(8, len(val))])
		} else {
			fmt.Fprintf(stdout, "  %-28s %s\n", env.label, val)
		}
	}

	fmt.Fprintf(stdout, "\n%s\n", ui.bold("registered tools"))
	groups := groupToolsByCategory(toolList)
	categories := sortedCategoryNames(groups)
	for _, category := range categories {
		entries := sortToolsForDisplay(groups[category])
		items := make([]string, 0, len(entries))
		for _, entry := range entries {
			items = append(items, shortToolName(entry))
		}
		icon, label := categoryMeta(category)
		fmt.Fprintf(stdout, "  %s %s (%d)\n", icon, ui.bold(label), len(entries))
		fmt.Fprintln(stdout, wrapJoinedItems(items, 76, "     "))
		fmt.Fprintln(stdout)
	}

	return 0
}

func printConfigUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  relay config")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
