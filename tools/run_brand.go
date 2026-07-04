// Package tools — run_brand is Agent 2: positioning, brand voice, messaging
// pillars, audience variants, tagline options.
//
// Reads pm_brief_for_agent1.md + 01_research.md from ./output/.
// Writes 02_brand_messaging.md atomically.
//
// No web search — brand work is grounded in the research dossier from Agent 1.
//
// Lock: 02_brand_messaging.md is guarded with a PID lockfile so concurrent
// invocations cannot corrupt the output mid-write.
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

func RunBrand(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	extraNotes := req.GetString("extra_notes", "")

	if err := state.AcquireLock("02_brand_messaging.md"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer state.ReleaseLock("02_brand_messaging.md")

	pmBrief, err := state.ReadOutput("pm_brief_for_agent1.md")
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("read pm brief: %v (run pm_plan first)", err),
		), nil
	}

	research, err := state.ReadOutput("01_research.md")
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("read research: %v (run run_research first)", err),
		), nil
	}

	system, err := loadPrompt("brand_agent.md")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Use ctxguard.Build so each labelled section is independently guarded
	// against the 120k-char window. Both sections are required — the dossier
	// can grow large after web search but we cannot make brand decisions
	// without it.
	contextBlock := ctxguard.Build([]ctxguard.Part{
		{Label: "PM Brief", Content: pmBrief, Required: true},
		{Label: "Research Dossier", Content: research, Required: true},
	})

	notesSection := ""
	if extraNotes != "" {
		notesSection = "\n\n## Changes Requested by Human Reviewer\n" +
			extraNotes + "\n\nIncorporate these changes."
	}

	user := fmt.Sprintf(`Produce the complete brand & messaging document.
Follow your system prompt spec exactly — all 5 sections, in order, with no extras.

%s%s

Produce the full brand document now.`,
		contextBlock,
		notesSection,
	)

	c := claude.New()
	result, err := c.Call(context.Background(), system, user)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM error: %v", err)), nil
	}

	if err := state.WriteOutput("02_brand_messaging.md", result); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("write error: %v", err)), nil
	}

	logger.Info("run_brand complete", "output", "02_brand_messaging.md")

	preview := result
	if len(preview) > 600 {
		preview = preview[:600] + "..."
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("Brand → ./output/02_brand_messaging.md\n\n%s", preview),
	), nil
}
