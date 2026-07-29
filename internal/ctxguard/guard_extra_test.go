// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Tamish Max
package ctxguard

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuard_JustOverThreshold(t *testing.T) {
	content := strings.Repeat("b", maxChars+10)
	result := Guard(content, "payload")
	assert.True(t, len(result) < len(content))
	assert.Contains(t, result, "[TRUNCATED: payload was ")
	assert.Contains(t, result, "chars. Showing first ")
}

func TestGuard_TruncationContainsLabelAndCounts(t *testing.T) {
	content := strings.Repeat("z", maxChars+500)
	result := Guard(content, "session-log")
	assert.Contains(t, result, "[TRUNCATED: session-log was ")
	assert.Contains(t, result, "Showing first 3000 chars")
}

func TestGuard_SummaryCharsExact(t *testing.T) {
	content := strings.Repeat("a", summaryChars) + "extra"
	result := Guard(content, "test")
	assert.Equal(t, summaryChars, len(strings.TrimSuffix(result, result[summaryChars:])))
}

func TestBuild_RequiredExceedsLimit_StillIncludesOptional(t *testing.T) {
	bigRequired := strings.Repeat("r", maxChars+50)
	parts := []Part{
		{Label: "BigReq", Content: bigRequired, Required: true},
		{Label: "SmallOpt", Content: "optional", Required: false},
	}
	result := Build(parts)
	assert.Contains(t, result, "## BigReq")
	assert.Contains(t, result, "[TRUNCATED: BigReq was ")
	assert.Contains(t, result, "## SmallOpt")
	assert.Less(t, len(result), maxChars)
}

func TestBuild_AllEmptyContents(t *testing.T) {
	parts := []Part{
		{Label: "Empty1", Content: "", Required: true},
		{Label: "Empty2", Content: "", Required: false},
	}
	result := Build(parts)
	assert.Contains(t, result, "## Empty1")
	assert.Contains(t, result, "## Empty2")
}

func TestBuild_DropsOptionalWhenTotalExceeds(t *testing.T) {
	smallReq := "req"
	bigOpt := strings.Repeat("o", maxChars+200)
	parts := []Part{
		{Label: "Req", Content: smallReq, Required: true},
		{Label: "BigOpt", Content: bigOpt, Required: false},
	}
	result := Build(parts)
	assert.Contains(t, result, "## Req")
	assert.Contains(t, result, "req")
	assert.Less(t, len(result), maxChars+200)
}

func TestBuild_SinglePartRequired(t *testing.T) {
	parts := []Part{
		{Label: "Solo", Content: "solo content", Required: true},
	}
	result := Build(parts)
	assert.Contains(t, result, "## Solo")
	assert.Contains(t, result, "solo content")
}

func TestBuild_SinglePartOptional(t *testing.T) {
	parts := []Part{
		{Label: "SoloOpt", Content: "solo opt", Required: false},
	}
	result := Build(parts)
	assert.Contains(t, result, "## SoloOpt")
	assert.Contains(t, result, "solo opt")
}

func TestGuard_LargeContent_HasCorrectPrefix(t *testing.T) {
	content := "HEADER_DATA\n" + strings.Repeat("x", maxChars+1000)
	result := Guard(content, "log")
	assert.True(t, strings.HasPrefix(result, "HEADER_DATA\n"))
}

func TestBuild_MultipleOptionalSeparatorCount(t *testing.T) {
	parts := []Part{
		{Label: "R", Content: "r", Required: true},
		{Label: "O1", Content: "o1", Required: false},
		{Label: "O2", Content: "o2", Required: false},
		{Label: "O3", Content: "o3", Required: false},
	}
	result := Build(parts)
	count := strings.Count(result, "---")
	assert.Equal(t, 3, count)
}
