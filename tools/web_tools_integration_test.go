package tools

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebToolsIntegrationFetchPreservesResponseBodies(t *testing.T) {
	t.Setenv("RELAY_SKIP_URL_VALIDATION", "1")

	responses := map[string]struct {
		contentType string
		body        string
	}{
		"/text": {
			contentType: "text/plain; charset=utf-8",
			body:        "plain response",
		},
		"/json": {
			contentType: "application/json",
			body:        `{"status":"ok","source":"relay"}`,
		},
		"/html": {
			contentType: "text/html; charset=utf-8",
			body:        "<!doctype html><title>Relay</title><main>web tool</main>",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, ok := responses[r.URL.Path]
		require.True(t, ok, "unexpected path %s", r.URL.Path)

		w.Header().Set("Content-Type", response.contentType)
		_, _ = w.Write([]byte(response.body))
	}))
	defer server.Close()

	for path, response := range responses {
		t.Run(strings.TrimPrefix(path, "/"), func(t *testing.T) {
			result := callTool(t, WebFetch, map[string]any{"url": server.URL + path})

			require.False(t, result.IsError)
			assert.Equal(t, response.body, resultText(t, result))
		})
	}
}

func TestWebToolsIntegrationFetchSendsMethodAndHeaders(t *testing.T) {
	t.Setenv("RELAY_SKIP_URL_VALIDATION", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, []string{"one", "two"}, r.Header.Values("X-Trace"))

		_, _ = w.Write([]byte(r.Method + ":" + strings.Join(r.Header.Values("X-Trace"), ",")))
	}))
	defer server.Close()

	result := callTool(t, WebFetch, map[string]any{
		"url":    server.URL,
		"method": http.MethodPost,
		"headers": map[string]any{
			"X-Trace": []any{"one", "two"},
		},
	})

	require.False(t, result.IsError)
	assert.Equal(t, "POST:one,two", resultText(t, result))
}

func TestWebToolsIntegrationFetchTruncatesLargeResponse(t *testing.T) {
	t.Setenv("RELAY_SKIP_URL_VALIDATION", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", int(maxToolFileBytes)+1)))
	}))
	defer server.Close()

	result := callTool(t, WebFetch, map[string]any{"url": server.URL})

	require.False(t, result.IsError)
	text := resultText(t, result)
	assert.Len(t, text, int(maxToolFileBytes)+len("\n\n[truncated at 1MB]"))
	assert.True(t, strings.HasSuffix(text, "\n\n[truncated at 1MB]"))
}

func TestWebToolsIntegrationStatusReportsHTTPStatus(t *testing.T) {
	t.Setenv("RELAY_SKIP_URL_VALIDATION", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(strings.Repeat("body", 2048)))
	}))
	defer server.Close()

	result := callTool(t, WebStatus, map[string]any{"url": server.URL})

	require.False(t, result.IsError)
	assert.True(t, strings.HasPrefix(resultText(t, result), "418 ("))
}

func TestWebToolsIntegrationStatusReportsConnectionErrors(t *testing.T) {
	t.Setenv("RELAY_SKIP_URL_VALIDATION", "1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	url := server.URL
	server.Close()

	result := callTool(t, WebStatus, map[string]any{"url": url})

	require.True(t, result.IsError)
	assert.Contains(t, resultText(t, result), "check url:")
}
