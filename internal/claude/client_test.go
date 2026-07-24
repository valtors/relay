package claude

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestEnvOr(t *testing.T) {
	t.Setenv("relay_TEST_STR", "override")
	if got := envOr("relay_TEST_STR", "default"); got != "override" {
		t.Fatalf("envOr = %q, want override", got)
	}
	if got := envOr("relay_TEST_MISSING_STR", "default"); got != "default" {
		t.Fatalf("envOr missing = %q, want default", got)
	}
}

func TestEnvInt(t *testing.T) {
	t.Setenv("relay_TEST_INT", "42")
	if got := envInt("relay_TEST_INT", 10); got != 42 {
		t.Fatalf("envInt = %d, want 42", got)
	}
	t.Setenv("relay_TEST_INT_BAD", "notanumber")
	if got := envInt("relay_TEST_INT_BAD", 10); got != 10 {
		t.Fatalf("envInt bad = %d, want 10", got)
	}
	t.Setenv("relay_TEST_INT_NEG", "-5")
	if got := envInt("relay_TEST_INT_NEG", 10); got != 10 {
		t.Fatalf("envInt negative = %d, want 10", got)
	}
	if got := envInt("relay_TEST_INT_MISSING", 7); got != 7 {
		t.Fatalf("envInt missing = %d, want 7", got)
	}
}

func TestEnvDuration(t *testing.T) {
	t.Setenv("relay_TEST_DUR", "60")
	if got := envDuration("relay_TEST_DUR", 30); got != 60 {
		t.Fatalf("envDuration = %v, want 60", got)
	}
	t.Setenv("relay_TEST_DUR_BAD", "abc")
	if got := envDuration("relay_TEST_DUR_BAD", 30); got != 30 {
		t.Fatalf("envDuration bad = %v, want 30", got)
	}
	if got := envDuration("relay_TEST_DUR_MISSING", 45); got != 45 {
		t.Fatalf("envDuration missing = %v, want 45", got)
	}
}

func TestErrorKind_Error(t *testing.T) {
	e := &Error{Kind: ErrRateLimit, Message: "too many requests"}
	want := "llm rate_limit: too many requests"
	if got := e.Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantKind  ErrorKind
		wantRetry bool
	}{
		{"429 rate limit", errors.New("status 429: rate limit exceeded"), ErrRateLimit, true},
		{"529 overloaded", errors.New("status 529: overloaded"), ErrRateLimit, true},
		{"deadline exceeded", context.DeadlineExceeded, ErrTimeout, true},
		{"generic api error", errors.New("something broke"), ErrAPI, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyError(tt.err)
			if got.Kind != tt.wantKind {
				t.Fatalf("Kind = %q, want %q", got.Kind, tt.wantKind)
			}
			if got.Retryable != tt.wantRetry {
				t.Fatalf("Retryable = %v, want %v", got.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
		wantVal string
	}{
		{"plain json", `{"queries": ["a", "b"]}`, false, "a"},
		{"json with markdown fence", "```json\n{\"queries\": [\"c\"]}\n```", false, "c"},
		{"json with bare fence", "```\n{\"queries\": [\"d\"]}\n```", false, "d"},
		{"json with whitespace", "  {\"queries\": [\"e\"]}  ", false, "e"},
		{"invalid json", `not json at all`, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dest struct {
				Queries []string `json:"queries"`
			}
			err := unmarshalJSON(tt.raw, &dest)
			if tt.wantErr {
				if err == nil {
					t.Fatal("unmarshalJSON err = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unmarshalJSON err = %v, want nil", err)
			}
			if len(dest.Queries) == 0 || dest.Queries[0] != tt.wantVal {
				t.Fatalf("Queries[0] = %q, want %q", dest.Queries, tt.wantVal)
			}
		})
	}
}

func TestFallbackSystemPrompt(t *testing.T) {
	base := "You are a research agent."
	got := fallbackSystemPrompt(base)
	if !strings.Contains(got, "FALLBACK MODE") {
		t.Fatal("fallbackSystemPrompt missing FALLBACK MODE marker")
	}
	if !strings.Contains(got, base) {
		t.Fatal("fallbackSystemPrompt missing original system prompt")
	}
	if !strings.Contains(got, "[unverified]") {
		t.Fatal("fallbackSystemPrompt missing [unverified] instruction")
	}
}

func TestMin(t *testing.T) {
	if got := min(3, 7); got != 3 {
		t.Fatalf("min(3,7) = %d, want 3", got)
	}
	if got := min(9, 2); got != 2 {
		t.Fatalf("min(9,2) = %d, want 2", got)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	calls := 0
	fn := func() (string, error) {
		calls++
		return "", &Error{Kind: ErrAPI, Message: "fatal", Retryable: false}
	}
	_, err := (&Client{}).withRetry(context.Background(), fn)
	if err == nil {
		t.Fatal("withRetry err = nil, want error")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (no retries for non-retryable)", calls)
	}
}

func TestWithRetry_RetryableErrorThenSuccess(t *testing.T) {
	calls := 0
	fn := func() (string, error) {
		calls++
		if calls < 2 {
			return "", &Error{Kind: ErrRateLimit, Message: "slow down", Retryable: true}
		}
		return "ok", nil
	}
	got, err := (&Client{}).withRetry(context.Background(), fn)
	if err != nil {
		t.Fatalf("withRetry err = %v, want nil", err)
	}
	if got != "ok" {
		t.Fatalf("withRetry = %q, want ok", got)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestWithRetry_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	fn := func() (string, error) {
		calls++
		return "", &Error{Kind: ErrRateLimit, Message: "slow down", Retryable: true}
	}
	_, err := (&Client{}).withRetry(ctx, fn)
	if err == nil {
		t.Fatal("withRetry err = nil, want context cancelled")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (should not retry after cancel)", calls)
	}
}

func TestCallJSON_NilError(t *testing.T) {
	c := &Client{}
	err := c.CallJSON(context.Background(), "system", "user", nil)
	if err == nil {
		t.Fatal("CallJSON err = nil, want error (no API key)")
	}
}

func TestModelDefaults(t *testing.T) {
	if Model == "" {
		t.Fatal("Model is empty")
	}
	if MaxTokens <= 0 {
		t.Fatalf("MaxTokens = %d, want positive", MaxTokens)
	}
	if Timeout <= 0 {
		t.Fatalf("Timeout = %v, want positive", Timeout)
	}
	if MaxRetries < 0 {
		t.Fatalf("MaxRetries = %d, want >= 0", MaxRetries)
	}
}

func TestCallWithSearch_NoAPIKey(t *testing.T) {
	c := &Client{}
	_, err := c.CallWithSearch(context.Background(), "system", "user")
	if err == nil {
		t.Fatal("CallWithSearch err = nil, want error (no API key)")
	}
}
