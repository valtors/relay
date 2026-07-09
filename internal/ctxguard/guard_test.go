package ctxguard

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuard_ShortContent(t *testing.T) {
	content := "short content"
	result := Guard(content, "test")
	assert.Equal(t, content, result)
}

func TestGuard_ExactThreshold(t *testing.T) {
	content := strings.Repeat("a", maxChars)
	result := Guard(content, "test")
	assert.Equal(t, content, result)
}

func TestGuard_OverThreshold(t *testing.T) {
	content := strings.Repeat("a", maxChars+1)
	result := Guard(content, "test")
	assert.True(t, strings.HasPrefix(result, strings.Repeat("a", summaryChars)))
	assert.Contains(t, result, "[TRUNCATED:")
	assert.Contains(t, result, "test")
	assert.Contains(t, result, "was 120001 chars")
}

func TestGuard_EmptyContent(t *testing.T) {
	result := Guard("", "test")
	assert.Equal(t, "", result)
}

func TestBuild_AllPartsBelowLimit(t *testing.T) {
	parts := []Part{
		{Label: "Required1", Content: "req content", Required: true},
		{Label: "Optional1", Content: "opt content", Required: false},
	}
	result := Build(parts)
	assert.Contains(t, result, "## Required1")
	assert.Contains(t, result, "req content")
	assert.Contains(t, result, "## Optional1")
	assert.Contains(t, result, "opt content")
	assert.Contains(t, result, "---")
}

func TestBuild_OnlyRequired(t *testing.T) {
	parts := []Part{
		{Label: "Req", Content: "required only", Required: true},
	}
	result := Build(parts)
	assert.Contains(t, result, "## Req")
	assert.Contains(t, result, "required only")
}

func TestBuild_OnlyOptional(t *testing.T) {
	parts := []Part{
		{Label: "Opt", Content: "optional only", Required: false},
	}
	result := Build(parts)
	assert.Contains(t, result, "## Opt")
	assert.Contains(t, result, "optional only")
}

func TestBuild_EmptyParts(t *testing.T) {
	result := Build([]Part{})
	assert.Equal(t, "\n\n---\n\n", result)
}

func TestBuild_TruncatesWhenTotalExceedsLimit(t *testing.T) {
	bigContent := strings.Repeat("x", maxChars+100)
	parts := []Part{
		{Label: "Req", Content: "small required", Required: true},
		{Label: "Opt", Content: bigContent, Required: false},
	}
	result := Build(parts)
	assert.Contains(t, result, "small required")
	assert.Contains(t, result, "[TRUNCATED: Opt was 120100 chars")
	assert.Less(t, len(result), maxChars)
}

func TestBuild_MultipleRequiredAndOptional(t *testing.T) {
	parts := []Part{
		{Label: "R1", Content: "r1", Required: true},
		{Label: "R2", Content: "r2", Required: true},
		{Label: "O1", Content: "o1", Required: false},
		{Label: "O2", Content: "o2", Required: false},
	}
	result := Build(parts)
	assert.Contains(t, result, "## R1")
	assert.Contains(t, result, "## R2")
	assert.Contains(t, result, "## O1")
	assert.Contains(t, result, "## O2")
}

func TestBuild_RequiredBeforeOptional(t *testing.T) {
	parts := []Part{
		{Label: "OptFirst", Content: "opt", Required: false},
		{Label: "ReqSecond", Content: "req", Required: true},
	}
	result := Build(parts)
	idxReq := strings.Index(result, "## ReqSecond")
	idxOpt := strings.Index(result, "## OptFirst")
	assert.True(t, idxReq >= 0 && idxOpt >= 0)
	assert.True(t, idxReq < idxOpt)
}
