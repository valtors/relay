package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chdirTemp swaps cwd to a fresh temp dir for the duration of the test so the
// state package's ./output/ writes don't pollute the repo.
func chdirTemp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	prev, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(prev) })
	return dir
}

// withStdin replaces os.Stdin with the given content (as a non-TTY pipe) for
// the duration of the test.
func withStdin(t *testing.T, content string) {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	_, err = w.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	prev := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = prev
		_ = r.Close()
	})
}

// makeReq builds a CallToolRequest carrying the given Arguments map, mimicking
// what the MCP server passes to a tool handler.
func makeReq(args map[string]any) mcp.CallToolRequest {
	var req mcp.CallToolRequest
	req.Params.Name = "request_approval"
	req.Params.Arguments = args
	return req
}

func parseResult(t *testing.T, res *mcp.CallToolResult) CheckpointResult {
	t.Helper()
	require.NotNil(t, res)
	require.NotEmpty(t, res.Content)
	tc, ok := res.Content[0].(mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T", res.Content[0])
	var cr CheckpointResult
	require.NoError(t, json.Unmarshal([]byte(tc.Text), &cr))
	return cr
}

// textOf extracts the .Text field from a tool result's first content block.
func textOf(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	require.NotNil(t, res)
	require.NotEmpty(t, res.Content)
	tc, ok := res.Content[0].(mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T", res.Content[0])
	return tc.Text
}

func TestRequestApproval_NonInteractiveAutoApproves(t *testing.T) {
	dir := chdirTemp(t)
	withStdin(t, "") // empty pipe — non-TTY, auto-approve path

	res, err := RequestApproval(t.Context(), makeReq(map[string]any{
		"checkpoint": "H1",
		"summary":    "Research complete",
		"questions":  []any{"OK to continue?"},
	}))
	require.NoError(t, err)

	cr := parseResult(t, res)
	assert.Equal(t, "approve", cr.Decision)
	assert.Contains(t, cr.Notes, "non-interactive")

	// Checkpoint file is written and includes the decision footer.
	body, err := os.ReadFile(filepath.Join(dir, "output", "checkpoint_H1.md"))
	require.NoError(t, err)
	assert.Contains(t, string(body), "# Checkpoint H1")
	assert.Contains(t, string(body), "Research complete")
	assert.Contains(t, string(body), "OK to continue?")
	assert.Contains(t, string(body), "## Decision")
	assert.Contains(t, string(body), "**APPROVE**")
}

func TestRequestApproval_MissingCheckpointReturnsError(t *testing.T) {
	chdirTemp(t)
	withStdin(t, "")

	res, err := RequestApproval(t.Context(), makeReq(map[string]any{
		"summary": "no checkpoint name",
	}))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsError, "expected IsError=true")
	tc, ok := res.Content[0].(mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, tc.Text, "checkpoint is required")
}

func TestRequestApproval_QuestionsAcceptsAnySlice(t *testing.T) {
	chdirTemp(t)
	withStdin(t, "")

	// MCP servers often deliver array args as []any after JSON unmarshaling.
	res, err := RequestApproval(t.Context(), makeReq(map[string]any{
		"checkpoint": "H2",
		"summary":    "brand done",
		"questions":  []any{"voice OK?", "pillars OK?"},
	}))
	require.NoError(t, err)
	parseResult(t, res) // shouldn't panic / shouldn't error

	body, err := os.ReadFile(filepath.Join(t.TempDir()+"/..", "checkpoint_H2.md"))
	if err != nil {
		// Recover: re-derive cwd based output dir
		body, err = os.ReadFile(filepath.Join("output", "checkpoint_H2.md"))
		require.NoError(t, err)
	}
	assert.Contains(t, string(body), "1. voice OK?")
	assert.Contains(t, string(body), "2. pillars OK?")
}

func TestRequestApproval_DefaultQuestionsWhenNoneProvided(t *testing.T) {
	chdirTemp(t)
	withStdin(t, "")

	res, err := RequestApproval(t.Context(), makeReq(map[string]any{
		"checkpoint": "H3",
		"summary":    "ux done",
	}))
	require.NoError(t, err)
	parseResult(t, res)

	body, err := os.ReadFile(filepath.Join("output", "checkpoint_H3.md"))
	require.NoError(t, err)
	assert.Contains(t, string(body), "Does this output match")
	assert.Contains(t, string(body), "Ready to proceed?")
}

func TestExtractQuestions(t *testing.T) {
	cases := []struct {
		name string
		args map[string]any
		want []string
	}{
		{"nil args", nil, nil},
		{"missing key", map[string]any{}, nil},
		{"[]any", map[string]any{"questions": []any{"a", "b"}}, []string{"a", "b"}},
		{"[]string", map[string]any{"questions": []string{"x", "y"}}, []string{"x", "y"}},
		{"mixed any drops non-strings", map[string]any{"questions": []any{"a", 42, "b"}}, []string{"a", "b"}},
		{"empty slice", map[string]any{"questions": []any{}}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractQuestions(makeReq(tc.args))
			if len(tc.want) == 0 {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	t.Setenv("TEST_KEY_INT", "")
	assert.Equal(t, 30, getEnvInt("TEST_KEY_INT", 30))

	t.Setenv("TEST_KEY_INT", "  90  ")
	assert.Equal(t, 90, getEnvInt("TEST_KEY_INT", 30))

	t.Setenv("TEST_KEY_INT", "not-a-number")
	assert.Equal(t, 30, getEnvInt("TEST_KEY_INT", 30))

	t.Setenv("TEST_KEY_INT", "-5")
	assert.Equal(t, 30, getEnvInt("TEST_KEY_INT", 30))
}

func TestBuildCheckpointDoc(t *testing.T) {
	doc := buildCheckpointDoc("H4", "summary text", []string{"q1", "q2", "q3"})
	assert.True(t, strings.HasPrefix(doc, "# Checkpoint H4"))
	assert.Contains(t, doc, "## PM Agent Summary\n\nsummary text")
	assert.Contains(t, doc, "1. q1")
	assert.Contains(t, doc, "2. q2")
	assert.Contains(t, doc, "3. q3")
}

func TestParseDecisionLine(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		decision string
		notes    string
	}{
		{"empty → approve", "", "approve", ""},
		{"approve lowercase", "approve", "approve", ""},
		{"approve uppercase", "APPROVE", "approve", ""},
		{"iterate alone (no notes)", "iterate", "iterate", ""},
		{"iterate with notes", "iterate fix the ICP", "iterate", "fix the ICP"},
		{"iterate uppercase + notes", "ITERATE redo brand voice", "iterate", "redo brand voice"},
		{"iterate with extra whitespace", "iterate    needs more competitors  ", "iterate", "needs more competitors"},
		{"ambiguous → approve with line as note", "looks fine to me", "approve", "looks fine to me"},
		{"iterate notes capped at maxIterateNotesLen", "iterate " + strings.Repeat("x", maxIterateNotesLen+500), "iterate", strings.Repeat("x", maxIterateNotesLen)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseDecisionLine(tc.input)
			require.NotNil(t, got)
			assert.Equal(t, tc.decision, got.Decision)
			assert.Equal(t, tc.notes, got.Notes)
		})
	}
}

func TestNotesOrNone(t *testing.T) {
	assert.Equal(t, "(none)", notesOrNone(""))
	assert.Equal(t, "fix it", notesOrNone("fix it"))
}
