package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/valtors/relay/internal/claude"
	"github.com/valtors/relay/internal/logger"
	"github.com/valtors/relay/internal/state"
)

func PMPlan(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	briefPath := req.GetString("brief_path", "")

	brief, err := state.ReadBrief(briefPath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	system, err := loadPrompt("pm_agent.md")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	user := fmt.Sprintf(`Read this product brief and write a tight one-page brief for Agent 1 (Research).

Include:
1. Exactly what to research (market size, key competitors, ICP demographics)
2. Hypotheses from the brief to validate or disprove
3. Constraints or existing knowledge to preserve
4. Required output format (market snapshot, ICP table, competitor table, strategic insights)

No padding. Every sentence must inform Agent 1.

## Product Brief
%s

Write the Agent 1 brief now, titled "Brief for Agent 1 — Research".`, brief)

	c := claude.New()
	result, err := c.Call(context.Background(), system, user)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("LLM error: %v", err)), nil
	}

	if err := state.WriteOutput("pm_brief_for_agent1.md", result); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("write error: %v", err)), nil
	}

	logger.Info("pm_plan complete", "output", "pm_brief_for_agent1.md")
	return mcp.NewToolResultText(
		fmt.Sprintf("PM brief → ./output/pm_brief_for_agent1.md\n\n%s", result),
	), nil
}
