package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"relay/internal/state"
)

func TestRunUX_MissingResearchReturnsToolError(t *testing.T) {
	chdirTemp(t)

	res, err := RunUX(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected tool error when 01_research.md is absent")
	assert.Contains(t, textOf(t, res), "research")
}

func TestRunUX_MissingBrandReturnsToolError(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("01_research.md", "# Research\nfindings"))

	res, err := RunUX(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected tool error when 02_brand_messaging.md is absent")
	assert.Contains(t, textOf(t, res), "brand")
}

func TestRunUX_WithUpstreamFilesReachesLLMCall(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("01_research.md", "# Research\nfindings"))
	require.NoError(t, state.WriteOutput("02_brand_messaging.md", "# Brand\nvoice"))
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	res, err := RunUX(context.Background(), makeReq(map[string]any{
		"extra_notes": "tighten the onboarding flow",
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "bogus API key should surface as LLM error")
	assert.Contains(t, textOf(t, res), "LLM error",
		"upstream files exist; failure must come from the LLM call, not earlier validation")
}

func TestRunUX_LockReleasedOnFailure(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("01_research.md", "# Research\nfindings"))
	require.NoError(t, state.WriteOutput("02_brand_messaging.md", "# Brand\nvoice"))
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	res1, _ := RunUX(context.Background(), makeReq(map[string]any{}))
	require.True(t, res1.IsError)

	res2, _ := RunUX(context.Background(), makeReq(map[string]any{}))
	require.True(t, res2.IsError, "second call should fail at the same LLM stage, not at lock acquisition")
	assert.Contains(t, textOf(t, res2), "LLM error",
		"if the lock had leaked, this would say 'lock held by'")
}

func TestUXAgentPromptEmbedded(t *testing.T) {
	body, err := loadPrompt("ux_agent.md")
	require.NoError(t, err)
	for _, marker := range []string{
		"You are Agent 3",
		"Core User Flows",
		"Screen List",
		"Wireframe Briefs",
		"Image-Prototype Prompts",
		"Interaction Notes",
	} {
		assert.True(t, strings.Contains(body, marker),
			"ux_agent.md must contain %q", marker)
	}
}
