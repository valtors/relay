// Package claude wraps the official Anthropic SDK with retry, timeout,
// structured error types, and a JSON extraction helper.
package claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"relay/internal/logger"
	"relay/internal/search"
)

var (
	Model      = envOr("relay_MODEL", "claude-opus-4-5")
	MaxTokens  = envInt("relay_MAX_TOKENS", 4096)
	Timeout    = envDuration("relay_TIMEOUT_SECONDS", 120) * time.Second
	MaxRetries = 3
	BaseDelay  = 2 * time.Second
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func envDuration(k string, defSec int) time.Duration {
	if v := os.Getenv(k); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			return time.Duration(n)
		}
	}
	return time.Duration(defSec)
}

// ErrorKind classifies LLM failures.
type ErrorKind string

const (
	ErrRateLimit ErrorKind = "rate_limit"
	ErrTimeout   ErrorKind = "timeout"
	ErrAPI       ErrorKind = "api_error"
	ErrParseJSON ErrorKind = "parse_error"
)

// Error is a structured LLM error with retry guidance.
type Error struct {
	Kind      ErrorKind
	Message   string
	Retryable bool
}

func (e *Error) Error() string {
	return fmt.Sprintf("llm %s: %s", e.Kind, e.Message)
}

// Client wraps the Anthropic SDK.
type Client struct {
	inner anthropic.Client
}

// New constructs a Client. Reads ANTHROPIC_API_KEY from the environment.
// If relay_ANTHROPIC_BASE_URL is set, the SDK is pointed at that URL instead
// of api.anthropic.com — useful for Anthropic-compatible proxies (mega, etc.).
//
// The SDK's per-request timeout is widened to match relay_TIMEOUT_SECONDS so
// large streaming responses (e.g. assemble_plan, which can take many minutes
// on slower models / proxies) aren't capped by the SDK's internal default.
func New() *Client {
	var opts []option.RequestOption
	if base := os.Getenv("relay_ANTHROPIC_BASE_URL"); base != "" {
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
		opts = append(opts, option.WithBaseURL(base))
		logger.Info("using custom anthropic base url", "url", base)
	}
	opts = append(opts, option.WithRequestTimeout(Timeout))
	return &Client{inner: anthropic.NewClient(opts...)}
}

// Call sends a standard completion request with retry + exponential backoff.
func (c *Client) Call(ctx context.Context, system, user string) (string, error) {
	logger.Info("calling claude", "model", Model)
	return c.withRetry(ctx, func() (string, error) {
		return c.doCall(ctx, system, user, nil)
	})
}

// CallWithSearch sends a completion request with a web_search tool enabled.
// Falls back to a plain Call if the response is empty (some Anthropic-compatible
// proxies, e.g. megallm, return 200 with empty content when they don't honour
// the web_search tool).
func (c *Client) CallWithSearch(ctx context.Context, system, user string) (string, error) {
	logger.Info("calling claude + web search", "model", Model)

	// Proxy mode: provider doesn't expose web_search server-side, so we run
	// our own DuckDuckGo search and inject the results into the user prompt.
	if os.Getenv("relay_ANTHROPIC_BASE_URL") != "" {
		return c.callWithDDG(ctx, system, user)
	}

	out, err := c.withRetry(ctx, func() (string, error) {
		tools := []anthropic.ToolUnionParam{{
			OfTool: &anthropic.ToolParam{
				Name:        "web_search",
				Description: anthropic.String("Search the web for current information"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"query": map[string]any{"type": "string"},
					},
					Required: []string{"query"},
				},
			},
		}}
		return c.doCall(ctx, system, user, tools)
	})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "" {
		logger.Info("web_search returned empty (proxy likely lacks support); falling back to DDG path")
		return c.callWithDDG(ctx, system, user)
	}
	return out, nil
}

// callWithDDG performs in-MCP web research:
//  1. Asks the LLM for a small set of search queries (cheap JSON call).
//  2. Runs them via DuckDuckGo's HTML SERP.
//  3. Injects the results into the user prompt and calls the LLM normally.
//
// If query generation or search both fail, falls back to the plain
// "training knowledge with [unverified] tags" prompt so the agent still
// produces something usable.
func (c *Client) callWithDDG(ctx context.Context, system, user string) (string, error) {
	queries, qErr := c.generateSearchQueries(ctx, system, user)
	if qErr != nil || len(queries) == 0 {
		logger.Info("query generation failed; using training-knowledge fallback", "err", qErr)
		return c.Call(ctx, fallbackSystemPrompt(system), user)
	}
	logger.Info("running DDG searches", "queries", len(queries))

	results := search.DDGMulti(ctx, queries, 5)
	if len(results) == 0 {
		logger.Info("DDG returned no results; using training-knowledge fallback")
		return c.Call(ctx, fallbackSystemPrompt(system), user)
	}
	logger.Info("DDG search complete", "results", len(results))

	augUser := user + "\n\n--- WEB_SEARCH_RESULTS ---\n" +
		"The following snippets were retrieved from DuckDuckGo to inform your response.\n" +
		"Cite specific sources by their bracketed number and the URL when you use a fact.\n\n" +
		search.FormatForPrompt(results) +
		"\n--- END WEB_SEARCH_RESULTS ---\n\n" +
		"Use the snippets above as primary evidence. If a needed datapoint is missing, " +
		"mark it [unverified] rather than fabricating a citation."
	return c.Call(ctx, system, augUser)
}

// generateSearchQueries asks the LLM to produce 3-5 focused web search queries
// for the given task. Uses a small token budget; falls back to nothing on parse error.
func (c *Client) generateSearchQueries(ctx context.Context, system, user string) ([]string, error) {
	prompt := "You will be given a research task. Produce 3 to 5 focused web search queries " +
		"that would gather the strongest evidence for the task. Output ONLY a JSON object " +
		"of the form {\"queries\": [\"...\", \"...\"]}. No prose, no markdown fences.\n\n" +
		"--- TASK SYSTEM PROMPT ---\n" + system +
		"\n\n--- TASK USER PROMPT ---\n" + user
	var dest struct {
		Queries []string `json:"queries"`
	}
	// Use a tighter call: small system, JSON-only.
	raw, err := c.Call(ctx, "You produce JSON-only responses. No markdown fences.", prompt)
	if err != nil {
		return nil, err
	}
	if err := unmarshalJSON(raw, &dest); err != nil {
		return nil, err
	}
	return dest.Queries, nil
}

func fallbackSystemPrompt(system string) string {
	return system + "\n\n" +
		"--- FALLBACK MODE ---\n" +
		"Web search is unavailable in this environment. Override any 'use web search' " +
		"or 'real current data only' instructions above. Produce the requested document " +
		"using your training knowledge. Mark time-sensitive figures with [unverified] " +
		"and add a one-line note at the top: '> Generated without live web search; " +
		"figures are illustrative and require verification.'"
}

// CallJSON calls Claude and unmarshals the response into dest.
// Retries once with a stricter prompt if JSON parsing fails.
func (c *Client) CallJSON(ctx context.Context, system, user string, dest any) error {
	raw, err := c.Call(ctx, system, user)
	if err != nil {
		return err
	}

	if parseErr := unmarshalJSON(raw, dest); parseErr == nil {
		return nil
	}

	logger.Warn("json parse failed — retrying with stricter prompt")
	raw2, err := c.Call(ctx, system,
		user+"\n\nIMPORTANT: Respond with ONLY valid JSON. No markdown fences. No preamble. No explanation.",
	)
	if err != nil {
		return err
	}

	if err := unmarshalJSON(raw2, dest); err != nil {
		return &Error{
			Kind:    ErrParseJSON,
			Message: fmt.Sprintf("invalid JSON after 2 attempts. Preview: %s", raw2[:min(200, len(raw2))]),
		}
	}
	return nil
}

func (c *Client) doCall(ctx context.Context, system, user string, tools []anthropic.ToolUnionParam) (string, error) {
	callCtx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(Model),
		MaxTokens: int64(MaxTokens),
		System: []anthropic.TextBlockParam{
			{Text: system},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(user)),
		},
	}
	if len(tools) > 0 {
		params.Tools = tools
	}

	// Stream the response. Streaming gives the SDK a per-chunk read deadline
	// (instead of a single end-to-end deadline), which is critical for slow
	// proxies and long generations like assemble_plan that can take 5-10+
	// minutes. It also lets us emit progress logs so the operator knows the
	// call is alive.
	stream := c.inner.Messages.NewStreaming(callCtx, params)
	defer stream.Close()

	msg := anthropic.Message{}
	lastLog := time.Now()
	for stream.Next() {
		event := stream.Current()
		if err := msg.Accumulate(event); err != nil {
			return "", classifyError(err)
		}
		if time.Since(lastLog) >= 30*time.Second {
			logger.Info("streaming...", "tokens_out_so_far", msg.Usage.OutputTokens)
			lastLog = time.Now()
		}
	}
	if err := stream.Err(); err != nil {
		return "", classifyError(err)
	}

	var sb strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}

	text := sb.String()
	logger.Info("response received",
		"chars", len(text),
		"tokens_out", msg.Usage.OutputTokens,
	)
	return text, nil
}

func (c *Client) withRetry(ctx context.Context, fn func() (string, error)) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= MaxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		var llmErr *Error
		if !errors.As(err, &llmErr) || !llmErr.Retryable || attempt == MaxRetries {
			return "", err
		}

		delay := time.Duration(math.Pow(2, float64(attempt))) * BaseDelay
		logger.Warn("llm error — retrying",
			"kind", llmErr.Kind,
			"attempt", attempt+1,
			"max", MaxRetries,
			"delay", delay,
		)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(delay):
		}
		lastErr = err
	}
	return "", lastErr
}

func classifyError(err error) *Error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "429") || strings.Contains(msg, "529") || strings.Contains(msg, "rate"):
		return &Error{Kind: ErrRateLimit, Message: msg, Retryable: true}
	case errors.Is(err, context.DeadlineExceeded):
		return &Error{Kind: ErrTimeout, Message: msg, Retryable: true}
	default:
		return &Error{Kind: ErrAPI, Message: msg, Retryable: false}
	}
}

func unmarshalJSON(raw string, dest any) error {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	return json.Unmarshal([]byte(clean), dest)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
