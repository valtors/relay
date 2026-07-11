package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valtors/relay/internal/state"
)

func TestAssemblePlan_TolerantOfMissingStageOutputs(t *testing.T) {
	chdirTemp(t)
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	res, err := AssemblePlan(context.Background(), makeReq(map[string]any{
		"product_name": "TestProd",
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "bogus API key should surface as LLM error")
	assert.Contains(t, textOf(t, res), "LLM error",
		"should reach the LLM call regardless of missing upstream files")
}

func TestAssemblePlan_DefaultsProductNameWhenBlank(t *testing.T) {
	chdirTemp(t)
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	res, _ := AssemblePlan(context.Background(), makeReq(map[string]any{
		"product_name": "   ",
	}))
	require.NotNil(t, res)
	require.True(t, res.IsError)
	assert.Contains(t, textOf(t, res), "LLM error")
}

func TestAssemblePlan_CollectsCheckpointFilesIntoAppendix(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("checkpoint_H1.md", "# H1\nresearch approved"))
	require.NoError(t, state.WriteOutput("checkpoint_H3.md", "# H3\nux approved with notes"))

	got := buildHumanNotesAppendix(nil)
	assert.Contains(t, got, "### H1")
	assert.Contains(t, got, "research approved")
	assert.Contains(t, got, "### H3")
	assert.Contains(t, got, "ux approved with notes")
	assert.NotContains(t, got, "### H2", "H2 has no file → must not appear")
	assert.NotContains(t, got, "### H4", "H4 has no file → must not appear")
}

func TestBuildHumanNotesAppendix_PrefersSessionMetaNotes(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.InitSession("brief.md"))
	require.NoError(t, state.SaveHumanNote("H1", "tighten ICP"))
	require.NoError(t, state.SaveHumanNote("H2", ""))
	require.NoError(t, state.WriteOutput("checkpoint_H1.md", "stale checkpoint copy"))

	meta, err := state.LoadSession()
	require.NoError(t, err)

	got := buildHumanNotesAppendix(meta)
	assert.Contains(t, got, "### H1\ntighten ICP")
	assert.Contains(t, got, "### H2\n(none)")
	assert.NotContains(t, got, "stale checkpoint copy", "session meta should be the primary source")
}

func TestAssemblePlan_PromptStructureBuildsCorrectly(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, state.WriteOutput("01_research.md", "# Research"))
	require.NoError(t, state.WriteOutput("02_brand_messaging.md", "# Brand"))
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")

	res, _ := AssemblePlan(context.Background(), makeReq(map[string]any{
		"product_name": "Acme Widget",
	}))
	require.NotNil(t, res)
	require.True(t, res.IsError)
	_, err := state.ReadOutput("final_product_plan.md")
	require.Error(t, err, "no output file should be written when LLM call fails")
}
