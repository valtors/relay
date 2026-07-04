// Package tools — assemble_plan stitches every stage output and the human-
// review artifacts into a single shippable document: final_product_plan.md.
//
// It deliberately does NOT bail on a missing stage file: a partial pipeline
// (e.g. the user only ran research + brand) should still produce a useful
// document with whatever data is available. Empty sections are tolerated.
package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"relay/internal/claude"
	"relay/internal/ctxguard"
	"relay/internal/logger"
	"relay/internal/state"
)

func AssemblePlan(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	productName := strings.TrimSpace(req.GetString("product_name", ""))
	if productName == "" {
		productName = "Product"
	}

	ctx := context.Background()

	sessionMeta, _ := state.LoadSession()
	briefPath := ""
	if sessionMeta != nil {
		briefPath = strings.TrimSpace(sessionMeta.BriefPath)
	}

	// Read every stage output. Errors are deliberately swallowed: assembly
	// runs even with missing pieces (the LLM is told to mark gaps explicitly).
	research, _ := state.ReadOutput("01_research.md")
	brand, _ := state.ReadOutput("02_brand_messaging.md")
	ux, _ := state.ReadOutput("03_ux.md")
	gtm, _ := state.ReadOutput("04_go_to_market.md")
	brief, _ := state.ReadBrief(briefPath)

	// Appendix A prefers canonical notes saved in session meta, with the older
	// checkpoint-file reconstruction retained as a fallback for legacy runs.
	humanNotes := buildHumanNotesAppendix(sessionMeta)

	system, err := loadPrompt("pm_agent.md")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// All source materials, labeled. ctxguard drops UX/GTM first if overflow.
	sources := ctxguard.Build([]ctxguard.Part{
		{Label: "Original Brief", Content: brief, Required: true},
		{Label: "Research", Content: research, Required: true},
		{Label: "Brand Messaging", Content: brand, Required: true},
		{Label: "UX", Content: ux, Required: false},
		{Label: "GTM", Content: gtm, Required: false},
	})

	// Multi-pass assembly. A single call (or even two) lets opus produce
	// some sections verbosely and then "naturally stop" before later
	// sections get written, even when well under max_tokens. Splitting
	// into four passes (one per section pair) forces each section to be
	// attempted independently and produces a reliably complete plan.
	c := claude.New()

	type pass struct {
		label  string
		header string
	}
	passes := []pass{
		{
			label: "1/4 (title + sections 0-1)",
			header: fmt.Sprintf(`

Produce ONLY the title and sections 0-1 of the final product plan, with this exact structure:

# %s — MVP Product Plan

## 0. Executive Summary
## 1. Product MVP Brief

Be comprehensive within these two sections. Do NOT include any other sections — they will be generated in later passes. End your response immediately after section 1.`, productName),
		},
		{
			label: "2/4 (sections 2-3)",
			header: `

Produce ONLY sections 2-3 of the final product plan, with this exact structure:

## 2. Research
## 3. Brand & Messaging

Be comprehensive within these two sections. Do NOT include the title or any other sections. Start your response with "## 2. Research" and end immediately after section 3.`,
		},
		{
			label: "3/4 (sections 4-5)",
			header: `

Produce ONLY sections 4-5 of the final product plan, with this exact structure:

## 4. UX & Prototype Prompts
## 5. Go-to-Market Plan

Be comprehensive within these two sections. Do NOT include the title or any other sections. Start your response with "## 4. UX & Prototype Prompts" and end immediately after section 5.`,
		},
		{
			label: "4/4 (sections 6-7 + Appendix B)",
			header: `

Produce ONLY sections 6, 7, and Appendix B of the final product plan, with this exact structure:

## 6. Launch Checklist
## 7. Open Risks & Next Steps
## Appendix B. References & Sources

Be comprehensive within these sections. Do NOT include the title, sections 0-5, or Appendix A — Appendix A is appended by the orchestrator from the human approval log. Start your response with "## 6. Launch Checklist".`,
		},
	}

	parts := make([]string, len(passes))
	for i, p := range passes {
		logger.Info("assemble_plan: pass " + p.label)
		out, err := c.Call(ctx, system, sources+p.header)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("LLM error (pass %s): %v", p.label, err)), nil
		}
		parts[i] = strings.TrimSpace(out)
	}

	// Stitch all four passes + Appendix A (human approval notes).
	final := strings.Join(parts, "\n\n") + "\n\n---\n\n" +
		"## Appendix A. Human Notes Log\n\n" + humanNotes

	if err := state.WriteOutput("final_product_plan.md", final); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("write error: %v", err)), nil
	}

	logger.Info("assemble_plan complete", "output", "final_product_plan.md")
	return mcp.NewToolResultText("✅ Final plan → ./output/final_product_plan.md"), nil
}

func buildHumanNotesAppendix(sessionMeta *state.SessionMeta) string {
	var sb strings.Builder

	if sessionMeta != nil && len(sessionMeta.HumanNotes) > 0 {
		for _, cp := range []string{"H1", "H2", "H3", "H4"} {
			note, ok := sessionMeta.HumanNotes[cp]
			if !ok {
				continue
			}
			note = strings.TrimSpace(note)
			if note == "" {
				note = "(none)"
			}
			fmt.Fprintf(&sb, "### %s\n%s\n\n", cp, note)
		}
		if sb.Len() > 0 {
			return sb.String()
		}
	}

	for _, cp := range []string{"H1", "H2", "H3", "H4"} {
		content, _ := state.ReadOutput(fmt.Sprintf("checkpoint_%s.md", cp))
		if strings.TrimSpace(content) != "" {
			fmt.Fprintf(&sb, "### %s\n%s\n\n", cp, content)
		}
	}
	if sb.Len() == 0 {
		return "(none recorded)\n"
	}
	return sb.String()
}
