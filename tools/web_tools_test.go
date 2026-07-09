package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebFetch(t *testing.T) {
	t.Setenv("RELAY_SKIP_URL_VALIDATION", "1")

	t.Run("get request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			_, _ = w.Write([]byte("hello from server"))
		}))
		defer server.Close()

		result := callTool(t, WebFetch, map[string]any{"url": server.URL})
		assert.False(t, result.IsError)
		assert.Equal(t, "hello from server", resultText(t, result))
	})

	t.Run("custom headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(r.Header.Get("X-Test-Token")))
		}))
		defer server.Close()

		result := callTool(t, WebFetch, map[string]any{
			"url": server.URL,
			"headers": map[string]any{
				"X-Test-Token": "relay-token",
			},
		})
		assert.False(t, result.IsError)
		assert.Equal(t, "relay-token", resultText(t, result))
	})

	t.Run("404 response", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		defer server.Close()

		result := callTool(t, WebFetch, map[string]any{"url": server.URL})
		assert.False(t, result.IsError)
		assert.Contains(t, resultText(t, result), "404 page not found")
	})

	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(200 * time.Millisecond):
				_, _ = w.Write([]byte("too slow"))
			}
		}))
		defer server.Close()

		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]any{"url": server.URL}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		result, err := WebFetch(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "fetch url:")
	})
}

func TestWebStatus(t *testing.T) {
	t.Setenv("RELAY_SKIP_URL_VALIDATION", "1")

	t.Run("reachable url", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		result := callTool(t, WebStatus, map[string]any{"url": server.URL})
		assert.False(t, result.IsError)
		assert.True(t, strings.HasPrefix(resultText(t, result), "200 ("))
	})

	t.Run("redirect handling", func(t *testing.T) {
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer target.Close()

		redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, target.URL, http.StatusFound)
		}))
		defer redirect.Close()

		result := callTool(t, WebStatus, map[string]any{"url": redirect.URL})
		assert.False(t, result.IsError)
		assert.True(t, strings.HasPrefix(resultText(t, result), "200 ("))
	})
}

func TestWebSSRFProtection(t *testing.T) {
	os.Unsetenv("RELAY_SKIP_URL_VALIDATION")

	t.Run("blocks localhost", func(t *testing.T) {
		result := callTool(t, WebFetch, map[string]any{"url": "http://127.0.0.1:9999"})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "non-public address")
	})

	t.Run("blocks cloud metadata", func(t *testing.T) {
		result := callTool(t, WebFetch, map[string]any{"url": "http://169.254.169.254/latest/meta-data/"})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "non-public address")
	})

	t.Run("blocks non-http scheme", func(t *testing.T) {
		result := callTool(t, WebFetch, map[string]any{"url": "file:///etc/passwd"})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "scheme must be")
	})

	t.Run("blocks private ip", func(t *testing.T) {
		result := callTool(t, WebFetch, map[string]any{"url": "http://10.0.0.1/"})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "non-public address")
	})
}
