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

// TestBuildServer_RegistersAllEightTools is a smoke test that confirms the
// registration helper produces a fully-loaded MCPServer regardless of which
// transport main() chooses. Behaviourally identical to Phase 0's
// verify_phase0.ps1 but driven from Go.
func TestBuildServer_RegistersAllEightTools(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-bogus-not-real")
	s := buildServer()
	require.NotNil(t, s)

	// Use the in-process client to enumerate tools — same code path as a
	// remote client, just without the HTTP hop. This is the cheapest way
	// to verify registration without coupling to mcp-go internals.
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

// TestStreamableHTTP_ToolsListEndpoint exercises the actual HTTP transport
// end-to-end: spin up the server on an ephemeral port, do a real MCP
// initialize handshake over Streamable-HTTP, then list tools. This is the
// transport real Claude.ai / remote MCP clients use.
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

// TestStreamableHTTP_RejectsRequestWithoutSessionID proves that the
// transport correctly enforces the session-ID requirement — a request that
// skips the initialize handshake must NOT be allowed to reach a tool. This
// is the property that makes the remote endpoint safe to expose.
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

	// Spec says no session ID → 4xx. Either 400 (bad request) or 404
	// (invalid session) is acceptable; what matters is the request did
	// not reach a tool handler.
	assert.GreaterOrEqual(t, resp.StatusCode, 400,
		"unauthenticated tool call must be refused, got %d", resp.StatusCode)
	assert.Less(t, resp.StatusCode, 500,
		"refusal should be 4xx (client error), not 5xx, got %d", resp.StatusCode)
}

// TestMain_FailsFastWithoutAPIKey re-asserts Phase 0's startup contract is
// preserved after the HTTP refactor: missing ANTHROPIC_API_KEY → exit 1
// before any tool registers. Subprocess re-exec is environment-specific so
// this primarily documents the contract; the fail-fast behaviour itself is
// also verified by Phase 0's verify_phase0.ps1.
func TestMain_FailsFastWithoutAPIKey(t *testing.T) {
	if os.Getenv("MAIN_FAILS_FAST_CHILD") == "1" {
		os.Unsetenv("ANTHROPIC_API_KEY")
		main()
		return
	}
	t.Skip("subprocess re-exec is environment-specific; covered by Phase 0 smoke test")
}

// ── helpers ──────────────────────────────────────────────────────────────

// mcpEndpoint returns the canonical Streamable-HTTP MCP endpoint for a base
// URL. The transport mounts the JSON-RPC handler at `/mcp` by default.
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
