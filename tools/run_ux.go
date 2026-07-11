package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/valtors/relay/internal/claude"
	"github.com/valtors/relay/internal/ctxguard"
	"github.com/valtors/relay/internal/logger"
	"github.com/valtors/relay/internal/state"
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
