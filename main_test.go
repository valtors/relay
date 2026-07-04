package main

import (
	"context"
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

func TestBuildServer_RegistersAllEightTools(t *testing.T) {
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
	assertHasAllEightTools(t, names)
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
	assertHasAllEightTools(t, names)
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

func assertHasAllEightTools(t *testing.T, got []string) {
	t.Helper()
	want := []string{
		"run_workflow", "pm_plan", "run_research", "run_brand", "run_ux",
		"run_gtm", "request_approval", "assemble_plan",
	}
	require.Len(t, got, len(want), "expected exactly %d tools, got %d: %v",
		len(want), len(got), got)
	for _, name := range want {
		assert.Contains(t, got, name, "missing tool %q in %v", name, got)
	}
}
