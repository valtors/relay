package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PMPlan with no brief in cwd → tool returns a descriptive error result
// (NOT a Go error — MCP errors are surfaced via CallToolResult.IsError).
func TestPMPlan_MissingBriefReturnsToolError(t *testing.T) {
	chdirTemp(t)

	res, err := PMPlan(t.Context(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.IsError, "expected IsError=true when product_brief.md is missing")

	tc := res.Content[0]
	require.NotNil(t, tc)
}

// Brief loading honours the brief_path argument.
func TestPMPlan_CustomBriefPathRead(t *testing.T) {
	dir := chdirTemp(t)
	custom := filepath.Join(dir, "custom-brief.md")
	require.NoError(t, os.WriteFile(custom, []byte("# Custom Brief\n\nbody"), 0o644))

	// We can't reach the LLM in tests (no API key, no network), but we can
	// verify the brief loader is wired by intercepting via state.ReadBrief
	// indirectly: we call the tool; if the brief load step fails it returns
	// a "not found" error. With a valid path, that branch is bypassed and
	// the next failure (LLM call) surfaces a different "LLM error: ..."
	// message — which is what we assert here.
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-invalid-on-purpose")

	res, err := PMPlan(t.Context(), makeReq(map[string]any{
		"brief_path": custom,
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected LLM call to fail with bogus key")

	// Should be the LLM-error path, NOT the brief-not-found path.
	body := textOf(t, res)
	assert.NotContains(t, body, "product_brief.md not found")
	assert.Contains(t, body, "LLM error")
}

// PM agent prompt is embedded and loadable.
func TestPMAgentPromptEmbedded(t *testing.T) {
	got, err := loadPrompt("pm_agent.md")
	require.NoError(t, err)
	assert.Contains(t, got, "You are the PM Agent")
	assert.Contains(t, got, "Responsibilities:")
}
