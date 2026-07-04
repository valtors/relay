package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"relay/internal/state"
)

func TestRunGTM_MissingResearchReturnsToolError(t *testing.T) {
	chdirTemp(t)

	res, err := RunGTM(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected tool error when 01_research.md is absent")
	assert.Contains(t, textOf(t, res), "research")
}

func TestRunGTM_MissingBrandReturnsToolError(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("01_research.md", "# Research\nfindings"))

	res, err := RunGTM(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected tool error when 02_brand_messaging.md is absent")
	assert.Contains(t, textOf(t, res), "brand")
}

func TestRunGTM_WithUpstreamFilesReachesLLMCall(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("01_research.md", "# Research\nfindings"))
	require.NoError(t, state.WriteOutput("02_brand_messaging.md", "# Brand\nvoice"))
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	res, err := RunGTM(context.Background(), makeReq(map[string]any{
		"extra_notes": "punchier post hooks",
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "bogus API key should surface as LLM error")
	// Either the social or b2b goroutine wins the race to fail — both are
	// valid signals that the parallel call wiring reached the LLM.
	body := textOf(t, res)
	assert.True(t,
		strings.Contains(body, "social agent:") || strings.Contains(body, "b2b agent:"),
		"expected social/b2b agent error, got: %s", body,
	)
}

func TestRunGTM_LockReleasedOnFailure(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("01_research.md", "# Research\nfindings"))
	require.NoError(t, state.WriteOutput("02_brand_messaging.md", "# Brand\nvoice"))
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	res1, _ := RunGTM(context.Background(), makeReq(map[string]any{}))
	require.True(t, res1.IsError)

	// If the lock had leaked from the first call the second call would error
	// with "lock held by ...". Instead it should reach the LLM stage again.
	res2, _ := RunGTM(context.Background(), makeReq(map[string]any{}))
	require.True(t, res2.IsError)
	body := textOf(t, res2)
	assert.False(t, strings.Contains(strings.ToLower(body), "lock held"),
		"lock should have been released on the first failure; got: %s", body)
}

func TestBuildGTMMerge_CombinesSocialB2BAndCalendar(t *testing.T) {
	merged := buildGTMMerge(
		"## 4a — Social Media Plan\nsocial body",
		"## 4b — B2B Outreach Plan\nb2b body",
	)
	for _, marker := range []string{
		"# 04 — Go-to-Market Plan",
		"## 4a — Social Media Plan",
		"social body",
		"## 4b — B2B Outreach Plan",
		"b2b body",
		"## Unified Launch Calendar",
		"| Week |",
		"Public launch",
	} {
		assert.True(t, strings.Contains(merged, marker),
			"merged GTM doc must contain %q", marker)
	}
}

func TestGTMAgentPromptEmbedded(t *testing.T) {
	body, err := loadPrompt("gtm_agent.md")
	require.NoError(t, err)
	for _, marker := range []string{
		"You are Agent 4",
		"Agent 4a",
		"Social Media Plan",
		"Agent 4b",
		"B2B Outreach Plan",
		"Day 0/3/7/14",
	} {
		assert.True(t, strings.Contains(body, marker),
			"gtm_agent.md must contain %q", marker)
	}
}
