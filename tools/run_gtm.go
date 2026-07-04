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

func RunGTM(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	extraNotes := req.GetString("extra_notes", "")

	if err := state.AcquireLock("04_go_to_market.md"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer state.ReleaseLock("04_go_to_market.md")

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

	system, err := loadPrompt("gtm_agent.md")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	contextBlock := ctxguard.Build([]ctxguard.Part{
		{Label: "Research Summary", Content: research, Required: true},
		{Label: "Brand Pillars", Content: brand, Required: true},
	})
	if extraNotes != "" {
		contextBlock += "\n\n## Iteration Notes\n" + extraNotes
	}

	type agentResult struct {
		output string
		err    error
	}
	socialCh := make(chan agentResult, 1)
	b2bCh := make(chan agentResult, 1)

	c := claude.New()
	ctx := context.Background()

	callAgent := func(label, intro string, ch chan<- agentResult) {
		defer func() {
			if r := recover(); r != nil {
				ch <- agentResult{err: fmt.Errorf("%s panic: %v", label, r)}
			}
		}()
		out, err := c.Call(ctx, system,
			intro+"\n\n"+contextBlock+
				"\n\nWrite the "+label+" plan with channels, sequences, and concrete copy. "+
				"Output as '## "+label+"'.",
		)
		ch <- agentResult{output: out, err: err}
	}

	go callAgent(
		"4a — Social Media Plan",
		"You are Agent 4a — Social Media GTM.",
		socialCh,
	)
	go callAgent(
		"4b — B2B Outreach Plan",
		"You are Agent 4b — B2B Outreach GTM.",
		b2bCh,
	)

	social := <-socialCh
	b2b := <-b2bCh

	if social.err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("social agent: %v", social.err)), nil
	}
	if b2b.err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("b2b agent: %v", b2b.err)), nil
	}

	merged := buildGTMMerge(social.output, b2b.output)
	if err := state.WriteOutput("04_go_to_market.md", merged); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("write error: %v", err)), nil
	}

	logger.Info("run_gtm complete (parallel)", "output", "04_go_to_market.md")

	preview := merged
	if len(preview) > 600 {
		preview = preview[:600] + "..."
	}
	return mcp.NewToolResultText(
		fmt.Sprintf("GTM (parallel) → ./output/04_go_to_market.md\n\n%s", preview),
	), nil
}

func buildGTMMerge(social, b2b string) string {
	return "# 04 — Go-to-Market Plan\n\n" + social +
		"\n\n---\n\n" + b2b +
		"\n\n---\n\n" +
		"## Unified Launch Calendar\n\n" +
		"| Week | Social Actions | B2B Actions | Milestone |\n" +
		"|------|---------------|-------------|----------|\n" +
		"| -2   | Teaser content, waitlist posts | Target list finalised | Waitlist open |\n" +
		"| -1   | Launch countdown, founder story | Cold sequence Day 0 | Press outreach |\n" +
		"| 0    | Launch post, demo video | Follow-up Day 3 | Public launch |\n" +
		"| +1   | User testimonials, feature posts | Day 7 follow-up | First conversions |\n" +
		"| +2   | Case study content | Day 14 re-engagement | Iteration round 1 |\n"
}
