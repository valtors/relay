package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valtors/relay/internal/state"
)

func TestRunBrand_MissingPMBriefReturnsToolError(t *testing.T) {
	chdirTemp(t)

	res, err := RunBrand(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected tool error when pm_brief is absent")
	assert.Contains(t, textOf(t, res), "pm brief")
}

func TestRunBrand_MissingResearchReturnsToolError(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("pm_brief_for_agent1.md", "# Brief\nbody"))

	res, err := RunBrand(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected tool error when 01_research.md is absent")
	assert.Contains(t, textOf(t, res), "research")
}

func TestRunBrand_WithUpstreamFilesReachesLLMCall(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("pm_brief_for_agent1.md", "# Brief\nbody"))
	require.NoError(t, state.WriteOutput("01_research.md", "# Research\nfindings"))
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	res, err := RunBrand(context.Background(), makeReq(map[string]any{
		"extra_notes": "be punchier",
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "bogus API key should surface as LLM error")
	assert.Contains(t, textOf(t, res), "LLM error",
		"upstream files exist; failure must come from the LLM call, not earlier validation")
}

func TestRunBrand_LockReleasedOnFailure(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("pm_brief_for_agent1.md", "# Brief\nbody"))
	require.NoError(t, state.WriteOutput("01_research.md", "# Research\nfindings"))
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	res1, _ := RunBrand(context.Background(), makeReq(map[string]any{}))
	require.True(t, res1.IsError)

	res2, _ := RunBrand(context.Background(), makeReq(map[string]any{}))
	require.True(t, res2.IsError, "second call should fail at the same LLM stage, not at lock acquisition")
	assert.Contains(t, textOf(t, res2), "LLM error",
		"if the lock had leaked, this would say 'lock held by'")
}

func TestBrandAgentPromptEmbedded(t *testing.T) {
	body, err := loadPrompt("brand_agent.md")
	require.NoError(t, err)
	for _, marker := range []string{
		"You are Agent 2",
		"Positioning Statement",
		"Brand Voice",
		"Messaging Pillars",
		"Audience Variants",
		"Tagline Options",
	} {
		assert.True(t, strings.Contains(body, marker),
			"brand_agent.md must contain %q", marker)
	}
}
