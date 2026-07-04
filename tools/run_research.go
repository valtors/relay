// Package tools — run_research is Agent 1: market snapshot, ICP classification,
// competitor analysis, strategic insights — backed by Claude's web_search tool.
//
// Reads pm_brief_for_agent1.md from ./output/ (produced by pm_plan).
// Writes 01_research.md atomically.
//
// Lock: 01_research.md is guarded with a PID lockfile so concurrent invocations
// (e.g. an accidental double-call from a flaky MCP client) cannot corrupt the
// output mid-write.
package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"relay/internal/claude"
	"relay/internal/ctxguard"
	"relay/internal/logger"
	"relay/internal/state"
)

func RunResearch(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	extraNotes := req.GetString("extra_notes", "")

	if err := state.AcquireLock("01_research.md"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer state.ReleaseLock("01_research.md")

	pmBrief, err := state.ReadOutput("pm_brief_for_agent1.md")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	system, err := loadPrompt("research_agent.md")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	notesSection := ""
	if extraNotes != "" {
		notesSection = "\n\n## Changes Requested by Human Reviewer\n" +
			extraNotes + "\n\nIncorporate these changes."
	}

	user := fmt.Sprintf(`Use web search to produce a complete research dossier.
Follow your system prompt spec exactly.

## Your Brief
%s%s

Produce the full research document now.`,
		ctxguard.Guard(pmBrief, "PM Brief"),
		notesSection,
	)

	c := claude.New()
	result, err := c.CallWithSearch(context.Background(), system, user)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM error: %v", err)), nil
	}

	if err := state.WriteOutput("01_research.md", result); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("write error: %v", err)), nil
	}

	logger.Info("run_research complete", "output", "01_research.md")

	preview := result
	if len(preview) > 600 {
		preview = preview[:600] + "..."
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("Research → ./output/01_research.md\n\n%s", preview),
	), nil
}
