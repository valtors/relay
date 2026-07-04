package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildServer_RegistersAllTools(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")
	s := buildServer()
	require.NotNil(t, s)

	c, err := client.NewInProcessClient(s)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = c.Initialize(ctx, mcp.InitializeRequest{})
	require.NoError(t, err, "in-process initialize handshake")

	resp, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	require.NoError(t, err)

	names := toolNames(resp)
	assertHasAllTools(t, names)
}

func TestStreamableHTTP_ToolsListEndpoint(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")
	s := buildServer()

	ts := server.NewTestStreamableHTTPServer(s)
	t.Cleanup(ts.Close)

	c, err := client.NewStreamableHttpClient(mcpEndpoint(ts.URL))
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, c.Start(ctx), "client start")

	_, err = c.Initialize(ctx, mcp.InitializeRequest{})
	require.NoError(t, err, "Streamable-HTTP initialize handshake")

	resp, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	require.NoError(t, err)

	names := toolNames(resp)
	assertHasAllTools(t, names)
}

func TestStreamableHTTP_RejectsRequestWithoutSessionID(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")
	s := buildServer()
	ts := server.NewTestStreamableHTTPServer(s)
	t.Cleanup(ts.Close)

	body := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`)
	req, err := http.NewRequest(http.MethodPost, mcpEndpoint(ts.URL), body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.GreaterOrEqual(t, resp.StatusCode, 400,
		"unauthenticated tool call must be refused, got %d", resp.StatusCode)
	assert.Less(t, resp.StatusCode, 500,
		"refusal should be 4xx (client error), not 5xx, got %d", resp.StatusCode)
}

func TestMain_FailsFastWithoutAPIKey(t *testing.T) {
	if os.Getenv("MAIN_FAILS_FAST_CHILD") == "1" {
		os.Unsetenv("ANTHROPIC_API_KEY")
		main()
		return
	}
	t.Skip("subprocess re-exec is environment-specific; covered by Phase 0 smoke test")
}

func TestRunCLI_Version(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runCLI([]string{"version"}, &stdout, &stderr)

	assert.Equal(t, 0, code)
	assert.Equal(t, Version+"\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestRunCLI_Status(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runCLI([]string{"status"}, &stdout, &stderr)

	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "relay v"+Version)
	assert.Contains(t, stdout.String(), "tools: 27 registered (5 categories)")
	assert.Contains(t, stdout.String(), "transport: stdio (default) | http (with --http)")
	assert.Empty(t, stderr.String())
}

func TestRunCLI_ToolsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runCLI([]string{"tools", "--json"}, &stdout, &stderr)

	require.Equal(t, 0, code)
	assert.Empty(t, stderr.String())

	var tools []toolInfo
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &tools))
	require.Len(t, tools, 27)
	assert.Equal(t, "data", tools[0].Category)
	assert.Equal(t, "data_csv_to_json", tools[0].Name)
	assert.Equal(t, "workflow", tools[len(tools)-1].Category)
}

func TestRunCLI_ToolsText(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runCLI([]string{"tools"}, &stdout, &stderr)

	assert.Equal(t, 0, code)
	assert.Empty(t, stderr.String())
	assert.Contains(t, stdout.String(), "relay tools (27 total)")
	assert.Contains(t, stdout.String(), "workflow (8)")
	assert.Contains(t, stdout.String(), "text (6)")
	assert.Contains(t, stdout.String(), "file (7)")
	assert.Contains(t, stdout.String(), "data (4)")
}

func TestRunCLI_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runCLI([]string{"help"}, &stdout, &stderr)

	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Contains(t, stdout.String(), "relay start [--http] [--addr :8080]")
	assert.Empty(t, stderr.String())
}

func TestRunCLI_UnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := runCLI([]string{"bogus"}, &stdout, &stderr)

	assert.Equal(t, 1, code)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), `relay: unknown command "bogus"`)
}

func TestTruncateDescription(t *testing.T) {
	got := truncateDescription("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 60)
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234...", got)
}

func mcpEndpoint(baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/mcp"
}

func toolNames(resp *mcp.ListToolsResult) []string {
	out := make([]string, 0, len(resp.Tools))
	for _, t := range resp.Tools {
		out = append(out, t.Name)
	}
	return out
}

func assertHasAllTools(t *testing.T, got []string) {
	t.Helper()
	want := []string{
		"run_workflow", "pm_plan", "run_research", "run_brand", "run_ux",
		"run_gtm", "request_approval", "assemble_plan",
		"text_word_count", "text_replace", "text_extract_regex",
		"text_base64_encode", "text_base64_decode", "text_md_to_html",
		"file_hash", "file_read", "file_write", "file_list",
		"file_size", "file_zip", "file_unzip",
		"data_json_format", "data_csv_to_json", "data_json_to_csv", "data_json_query",
		"web_fetch", "web_status",
	}
	require.Len(t, got, len(want), "expected exactly %d tools, got %d: %v",
		len(want), len(got), got)
	for _, name := range want {
		assert.Contains(t, got, name, "missing tool %q in %v", name, got)
	}
}
