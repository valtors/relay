package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunResearch_MissingPMBriefReturnsToolError(t *testing.T) {
	chdirTemp(t)

	res, err := RunResearch(t.Context(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.IsError)

	body := textOf(t, res)
	assert.Contains(t, body, "pm_brief_for_agent1.md")
	assert.Contains(t, body, "run the prior stage first")
}

func TestRunResearch_WithBriefReachesLLMCall(t *testing.T) {
	dir := chdirTemp(t)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "output"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "output", "pm_brief_for_agent1.md"),
		[]byte("# Brief for Agent 1\n\nResearch X, Y, Z."),
		0o644,
	))
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-invalid")

	res, err := RunResearch(t.Context(), makeReq(map[string]any{
		"extra_notes": "make sure to include EU market data",
	}))
	require.NoError(t, err)
	require.True(t, res.IsError)

	body := textOf(t, res)
	assert.Contains(t, body, "LLM error", "should fail at LLM call, not earlier")
}

func TestRunResearch_LockReleasedOnFailure(t *testing.T) {
	dir := chdirTemp(t)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "output"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "output", "pm_brief_for_agent1.md"),
		[]byte("brief"),
		0o644,
	))
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-invalid")

	res1, err := RunResearch(t.Context(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.True(t, res1.IsError)

	_, statErr := os.Stat(filepath.Join(dir, "output", "01_research.md.lock"))
	assert.True(t, os.IsNotExist(statErr), "lock should be released after failure")

	res2, err := RunResearch(t.Context(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.True(t, res2.IsError)
	assert.NotContains(t, textOf(t, res2), "is locked")
}

func TestResearchAgentPromptEmbedded(t *testing.T) {
	got, err := loadPrompt("research_agent.md")
	require.NoError(t, err)
	assert.Contains(t, got, "You are Agent 1 — Research")
	assert.Contains(t, got, "## 1. Market Snapshot")
	assert.Contains(t, got, "## 4. Strategic Insights")
}
