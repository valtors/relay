// Package tools — run_ux is Agent 3: user flows, screen list, wireframe briefs,
// image-prototype prompts, and interaction notes.
//
// Reads 01_research.md + 02_brand_messaging.md from ./output/.
// Writes 03_ux.md atomically.
//
// No web search — UX work is grounded in the research dossier (for ICPs/needs)
// and the brand document (for voice and visual tone in the prototype prompts).
//
// Lock: 03_ux.md is guarded with a PID lockfile so concurrent invocations
// cannot corrupt the output mid-write.
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

func RunUX(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	extraNotes := req.GetString("extra_notes", "")

	if err := state.AcquireLock("03_ux.md"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer state.ReleaseLock("03_ux.md")

	research, err := state.ReadOutput("01_research.md")
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("read research: %v (run run_research first)", err),
		), nil
	}

	brand, err := state.ReadOutput("02_brand_messaging.md")
	if err != nil {
		return mcp.NewToolResultError(
			fmt.Sprintf("read brand: %v (run run_brand first)", err),
		), nil
	}

	system, err := loadPrompt("ux_agent.md")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// ctxguard.Build labels each upstream so the LLM can cite them by name,
	// and applies the per-part 120k-char guard. Both required: UX without
	// either ICP data or brand voice would just be generic wireframes.
	contextBlock := ctxguard.Build([]ctxguard.Part{
		{Label: "Research Dossier", Content: research, Required: true},
		{Label: "Brand & Messaging", Content: brand, Required: true},
	})

	notesSection := ""
	if extraNotes != "" {
		notesSection = "\n\n## Changes Requested by Human Reviewer\n" +
			extraNotes + "\n\nIncorporate these changes."
	}

	user := fmt.Sprintf(`Produce the complete UX document.
Follow your system prompt spec exactly — all 5 sections, in order, with no extras.
Image-prototype prompts must reference the brand voice adjectives by name.

%s%s

Produce the full UX document now.`,
		contextBlock,
		notesSection,
	)

	c := claude.New()
	result, err := c.Call(context.Background(), system, user)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM error: %v", err)), nil
	}

	if err := state.WriteOutput("03_ux.md", result); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("write error: %v", err)), nil
	}

	logger.Info("run_ux complete", "output", "03_ux.md")

	preview := result
	if len(preview) > 600 {
		preview = preview[:600] + "..."
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("UX → ./output/03_ux.md\n\n%s", preview),
	), nil
}
