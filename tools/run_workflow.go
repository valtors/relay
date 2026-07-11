package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/valtors/relay/internal/claude"
	"github.com/valtors/relay/internal/ctxguard"
	"github.com/valtors/relay/internal/logger"
	"github.com/valtors/relay/internal/state"
)

func RunWorkflow(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	briefPath := strings.TrimSpace(req.GetString("brief_path", ""))
	ctx := context.Background()

	sessionExists := state.OutputExists(".session.meta.json")
	resume := getOptionalBoolArg(req, "resume", sessionExists)

	switch {
	case !sessionExists:
		if err := state.InitSession(briefPath); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("init session: %v", err)), nil
		}
		logger.Info("new session started")

	case !resume:
		if err := state.InitSession(briefPath); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("reset session: %v", err)), nil
		}
		logger.Info("starting fresh session")

	default:
		meta, err := state.LoadSession()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("load session: %v", err)), nil
		}
		if briefPath == "" {
			briefPath = strings.TrimSpace(meta.BriefPath)
		}
		logger.Info("resuming existing session")
	}

	if !state.IsStageComplete(state.StageResearch) {
		logger.Info("stage 1/4: research")

		if err := runTool(PMPlan, ctx, map[string]any{"brief_path": briefPath}); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if err := runStage(ctx, stageConfig{
			key:        state.StageResearch,
			checkpoint: "H1",
			outputFile: "01_research.md",
			run: func(notes string) error {
				args := map[string]any{}
				if notes != "" {
					args["extra_notes"] = notes
				}
				return runTool(RunResearch, ctx, args)
			},
		}); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	if !state.IsStageComplete(state.StageBrand) {
		logger.Info("stage 2/4: brand")
		if err := runStage(ctx, stageConfig{
			key:        state.StageBrand,
			checkpoint: "H2",
			outputFile: "02_brand_messaging.md",
			run: func(notes string) error {
				return runTool(RunBrand, ctx, map[string]any{"extra_notes": notes})
			},
		}); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	if !state.IsStageComplete(state.StageUX) {
		logger.Info("stage 3/4: ux")
		if err := runStage(ctx, stageConfig{
			key:        state.StageUX,
			checkpoint: "H3",
			outputFile: "03_ux.md",
			run: func(notes string) error {
				return runTool(RunUX, ctx, map[string]any{"extra_notes": notes})
			},
		}); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	if !state.IsStageComplete(state.StageGTM) {
		logger.Info("stage 4/4: gtm")
		if err := runStage(ctx, stageConfig{
			key:        state.StageGTM,
			checkpoint: "H4",
			outputFile: "04_go_to_market.md",
			run: func(notes string) error {
				return runTool(RunGTM, ctx, map[string]any{"extra_notes": notes})
			},
		}); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	logger.Info("assembling final plan")
	if err := runTool(AssemblePlan, ctx, map[string]any{}); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("✅ Pipeline complete → ./output/final_product_plan.md"), nil
}

type stageConfig struct {
	key        state.StageKey
	checkpoint string
	outputFile string
	run        func(notes string) error
}

func runStage(ctx context.Context, cfg stageConfig) error {
	maxIter := getEnvInt("MAX_ITERATIONS_PER_STAGE", 5)
	if maxIter <= 0 {
		maxIter = 5
	}
	notes := ""

	for i := 0; i < maxIter; i++ {
		if err := cfg.run(notes); err != nil {
			return fmt.Errorf("%s run error: %w", cfg.key, err)
		}

		output, err := state.ReadOutput(cfg.outputFile)
		if err != nil {
			return fmt.Errorf("read %s: %w", cfg.outputFile, err)
		}

		summary, questions, err := pmSummarize(ctx, output, cfg.checkpoint)
		if err != nil {
			logger.Warn("summarize failed — using fallback", "err", err)
			summary = output[:minInt(500, len(output))]
			questions = []string{
				"Does this output match your expectations?",
				"Any areas to improve?",
				"Ready to proceed?",
			}
		}

		result, err := callApproval(ctx, cfg.checkpoint, summary, questions)
		if err != nil {
			return fmt.Errorf("approval error: %w", err)
		}

		if result.Decision == "approve" {
			if err := state.MarkStageComplete(cfg.key); err != nil {
				return fmt.Errorf("persist %s completion: %w", cfg.key, err)
			}
			return nil
		}

		notes = result.Notes
		n, err := state.IncrementIteration(cfg.key)
		if err != nil {
			return fmt.Errorf("persist %s iteration: %w", cfg.key, err)
		}
		logger.Info("iterating", "stage", cfg.key, "iteration", n, "max", maxIter)

		if i == maxIter-1 {
			logger.Warn("max iterations reached — proceeding with last output",
				"stage", cfg.key)
			if err := state.MarkStageComplete(cfg.key); err != nil {
				return fmt.Errorf("persist %s completion: %w", cfg.key, err)
			}
			return nil
		}
	}
	return nil
}

type summaryResponse struct {
	Summary   string   `json:"summary"`
	Questions []string `json:"questions"`
}

func pmSummarize(ctx context.Context, output, checkpoint string) (string, []string, error) {
	system, err := loadPrompt("pm_agent.md")
	if err != nil {
		return "", nil, err
	}

	user := fmt.Sprintf(`Checkpoint %s: read this stage output and write:
1. A 5-10 bullet executive summary for the human reviewer
2. Exactly 3 specific, answerable questions

Return ONLY valid JSON: {"summary": "...", "questions": ["...", "...", "..."]}

Stage output:
%s`, checkpoint, ctxguard.Guard(output, "Stage Output"))

	c := claude.New()
	var resp summaryResponse
	if err := c.CallJSON(ctx, system, user, &resp); err != nil {
		return "", nil, err
	}
	return resp.Summary, resp.Questions, nil
}

func runTool(
	fn func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error),
	ctx context.Context,
	args map[string]any,
) error {
	result, err := fn(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: args},
	})
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("tool returned nil result")
	}
	if result.IsError {
		if len(result.Content) > 0 {
			if t, ok := result.Content[0].(mcp.TextContent); ok {
				return fmt.Errorf("%s", t.Text)
			}
		}
		return fmt.Errorf("tool returned error")
	}
	return nil
}

func callApproval(ctx context.Context, checkpoint, summary string, questions []string) (*CheckpointResult, error) {
	qs := make([]any, len(questions))
	for i, q := range questions {
		qs[i] = q
	}

	result, err := RequestApproval(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{
			"checkpoint": checkpoint,
			"summary":    summary,
			"questions":  qs,
		}},
	})
	if err != nil {
		return nil, err
	}
	if result == nil || len(result.Content) == 0 {
		return nil, fmt.Errorf("approval returned empty content")
	}
	if result.IsError {
		if t, ok := result.Content[0].(mcp.TextContent); ok {
			return nil, fmt.Errorf("approval: %s", t.Text)
		}
		return nil, fmt.Errorf("approval returned tool error")
	}
	t, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		return nil, fmt.Errorf("approval returned non-text content")
	}

	var r CheckpointResult
	if err := json.Unmarshal([]byte(t.Text), &r); err != nil {
		return nil, fmt.Errorf("parse approval result: %w", err)
	}
	return &r, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getOptionalBoolArg(req mcp.CallToolRequest, key string, def bool) bool {
	args := req.GetArguments()
	if args == nil {
		return def
	}
	raw, ok := args[key]
	if !ok {
		return def
	}
	b, ok := raw.(bool)
	if !ok {
		return def
	}
	return b
}
