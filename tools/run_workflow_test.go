package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"relay/internal/state"
)

// ── Tests for the small in-process plumbing (runTool / callApproval) ──────

func TestRunTool_PropagatesIsErrorAsGoError(t *testing.T) {
	failing := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultError("synthetic failure"), nil
	}
	err := runTool(failing, context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "synthetic failure")
}

func TestRunTool_HappyPathReturnsNil(t *testing.T) {
	ok := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("done"), nil
	}
	require.NoError(t, runTool(ok, context.Background(), nil))
}

func TestRunTool_NilResultIsError(t *testing.T) {
	bad := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, nil
	}
	err := runTool(bad, context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil result")
}

func TestRunTool_GoErrorPassesThrough(t *testing.T) {
	bad := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, assertErr("boom")
	}
	err := runTool(bad, context.Background(), nil)
	require.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

type assertErr string

func (e assertErr) Error() string { return string(e) }

// ── Tests for callApproval (parses RequestApproval's JSON output) ────────

func TestCallApproval_AutoApprovesOnNonTTYStdin(t *testing.T) {
	// On a non-TTY stdin RequestApproval auto-approves and writes a checkpoint
	// document to ./output/. Use chdirTemp so we don't touch the real repo.
	chdirTemp(t)
	withStdin(t, "") // pipe → non-TTY → auto-approve path inside waitForDecision

	cr, err := callApproval(context.Background(), "H1", "summary", []string{"q1", "q2", "q3"})
	require.NoError(t, err)
	require.NotNil(t, cr)
	assert.Equal(t, "approve", cr.Decision)
	assert.Contains(t, cr.Notes, "auto-approved")
}

// ── Tests for runStage orchestration ─────────────────────────────────────

// fakeAgent records the `notes` arg of every invocation and writes the i-th
// element of `outputs` into the stage output file each time, simulating an
// agent that produces fresh content per iteration.
type fakeAgent struct {
	outputFile string
	runCalls   []string
	outputs    []string
}

func (f *fakeAgent) run(notes string) error {
	f.runCalls = append(f.runCalls, notes)
	idx := len(f.runCalls) - 1
	if idx >= len(f.outputs) {
		idx = len(f.outputs) - 1
	}
	return state.WriteOutput(f.outputFile, f.outputs[idx])
}

func TestRunStage_HappyPathApprovesOnFirstIteration(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.InitSession("brief.md"))
	// Non-TTY stdin → request_approval auto-approves on every call.
	withStdin(t, "")
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	// pmSummarize will fail without a real API key; runStage falls back to a
	// synthetic summary + 3 default questions and proceeds. That fallback
	// path is itself part of the contract we want to verify works.
	f := &fakeAgent{
		outputFile: "01_research.md",
		outputs:    []string{"# Research dossier\n\nlots of content"},
	}

	err := runStage(context.Background(), stageConfig{
		key:        state.StageResearch,
		checkpoint: "H1",
		outputFile: f.outputFile,
		run:        f.run,
	})
	require.NoError(t, err)
	assert.True(t, state.IsStageComplete(state.StageResearch),
		"stage should be marked complete after approve")
	assert.Len(t, f.runCalls, 1, "agent should run exactly once on first-shot approval")
	assert.Equal(t, "", f.runCalls[0], "first iteration carries empty notes")
}

func TestRunStage_AgentErrorPropagates(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.InitSession("brief.md"))

	failing := func(_ string) error { return assertErr("agent blew up") }

	err := runStage(context.Background(), stageConfig{
		key:        state.StageUX,
		checkpoint: "H3",
		outputFile: "03_ux.md",
		run:        failing,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent blew up")
	assert.False(t, state.IsStageComplete(state.StageUX),
		"failed stage must not be marked complete")
}

func TestRunStage_MissingOutputFileIsError(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.InitSession("brief.md"))

	// Agent "succeeds" but never writes the expected output file.
	noop := func(_ string) error { return nil }

	err := runStage(context.Background(), stageConfig{
		key:        state.StageGTM,
		checkpoint: "H4",
		outputFile: "04_go_to_market.md",
		run:        noop,
	})
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "read 04_go_to_market.md")
}

// ── Tests for RunWorkflow crash-resume ────────────────────────────────────

func TestRunWorkflow_ResumesByCheckingMeta(t *testing.T) {
	chdirTemp(t)
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	// Pre-create a session meta with all four stages already complete so the
	// orchestrator should skip every stage and only call AssemblePlan.
	require.NoError(t, state.InitSession("nonexistent_brief.md"))
	require.NoError(t, state.MarkStageComplete(state.StageResearch))
	require.NoError(t, state.MarkStageComplete(state.StageBrand))
	require.NoError(t, state.MarkStageComplete(state.StageUX))
	require.NoError(t, state.MarkStageComplete(state.StageGTM))

	// AssemblePlan is now real and will attempt an LLM call. With a bogus
	// key it surfaces an LLM error — that's our signal that the orchestrator
	// correctly *skipped* every agent stage and went straight to assembly.
	// (If any agent stage had run, we would see an agent-specific error
	// message instead, e.g. "read pm brief".)
	res, err := RunWorkflow(context.Background(), makeReq(map[string]any{
		"brief_path": "nonexistent_brief.md",
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "AssemblePlan with bogus key should surface LLM error")
	body := textOf(t, res)
	assert.Contains(t, body, "LLM error",
		"orchestrator should skip all agent stages and fail at assembly's LLM call")
	assert.NotContains(t, body, "pm brief",
		"no agent stage should have run if all stages were pre-marked complete")
	assert.NotContains(t, body, "research:",
		"no agent stage should have run if all stages were pre-marked complete")
}

func TestRunWorkflow_FailsFastWhenBriefMissingOnFreshSession(t *testing.T) {
	chdirTemp(t)
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	// No session meta and no brief file — first thing run_workflow does for a
	// new session is run pm_plan, which will fail to read the brief.
	res, err := RunWorkflow(context.Background(), makeReq(map[string]any{
		"brief_path": "missing_brief.md",
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected tool error for missing brief")

	// Session meta should still have been created (init happens before pm_plan).
	dir, err := state.OutputDir()
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, ".session.meta.json"))
	require.NoError(t, err, "session meta should be initialised before pm_plan runs")
}

func TestRunWorkflow_StageMetaIsPersisted(t *testing.T) {
	chdirTemp(t)

	require.NoError(t, state.InitSession("brief.md"))
	require.NoError(t, state.MarkStageComplete(state.StageResearch))

	// Round-trip the on-disk meta to confirm completion is durable across
	// process restarts (the property crash-resume relies on).
	dir, err := state.OutputDir()
	require.NoError(t, err)
	raw, err := os.ReadFile(filepath.Join(dir, ".session.meta.json"))
	require.NoError(t, err)

	var meta map[string]any
	require.NoError(t, json.Unmarshal(raw, &meta))
	completed, ok := meta["completedStages"].([]any)
	require.True(t, ok, "completedStages key missing or wrong type in %s", string(raw))
	assert.Contains(t, completed, "research", "research should be in completedStages on disk")
}

func TestRunWorkflow_ResumeFalseResetsExistingSession(t *testing.T) {
	chdirTemp(t)
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	require.NoError(t, state.InitSession("old_brief.md"))
	require.NoError(t, state.MarkStageComplete(state.StageResearch))
	require.NoError(t, state.SaveHumanNote("H1", "old note"))

	res, err := RunWorkflow(context.Background(), makeReq(map[string]any{
		"brief_path": "missing_brief.md",
		"resume":     false,
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected tool error for missing brief on fresh restart")

	body := textOf(t, res)
	assert.Contains(t, body, "missing_brief.md")
	assert.NotContains(t, body, "LLM error", "fresh run should restart at pm_plan, not jump to assembly")

	meta, err := state.LoadSession()
	require.NoError(t, err)
	assert.Equal(t, "missing_brief.md", meta.BriefPath)
	assert.Empty(t, meta.CompletedStages, "fresh run should clear completed stages")
	assert.Empty(t, meta.HumanNotes, "fresh run should clear prior reviewer notes")
}
