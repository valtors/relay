package tools

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextWordCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		text    string
		want    string
		isError bool
	}{
		{name: "empty string", text: "", want: "text is required", isError: true},
		{name: "one word", text: "hello", want: "1"},
		{name: "multiple words", text: "hello world from relay", want: "4"},
		{name: "unicode", text: "नमस्ते दुनिया", want: "2"},
		{name: "newlines", text: "one\ntwo\nthree", want: "3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, TextWordCount, map[string]any{"text": tc.text})
			assert.Equal(t, tc.isError, result.IsError)
			assert.Equal(t, tc.want, resultText(t, result))
		})
	}
}

func TestTextReplace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    string
		isError bool
	}{
		{
			name: "basic replace",
			args: map[string]any{"text": "hello world", "find": "world", "replace": "relay"},
			want: "hello relay",
		},
		{
			name: "replace all",
			args: map[string]any{"text": "go go go", "find": "go", "replace": "run", "all": true},
			want: "run run run",
		},
		{
			name: "replace first only",
			args: map[string]any{"text": "go go go", "find": "go", "replace": "run", "all": false},
			want: "run go go",
		},
		{
			name: "no match found",
			args: map[string]any{"text": "hello", "find": "xyz", "replace": "abc"},
			want: "hello",
		},
		{
			name:    "empty text",
			args:    map[string]any{"text": "", "find": "a", "replace": "b"},
			want:    "text is required",
			isError: true,
		},
		{
			name:    "empty find",
			args:    map[string]any{"text": "hello", "find": "", "replace": "b"},
			want:    "find is required",
			isError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, TextReplace, tc.args)
			assert.Equal(t, tc.isError, result.IsError)
			assert.Equal(t, tc.want, resultText(t, result))
		})
	}
}

func TestTextExtractRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    []string
		wantErr string
	}{
		{
			name: "simple pattern",
			args: map[string]any{"text": "cat bat rat", "pattern": "b\\w+"},
			want: []string{"bat"},
		},
		{
			name: "groups",
			args: map[string]any{"text": "IDs: 12, 34", "pattern": "(\\d+)"},
			want: []string{"12", "34"},
		},
		{
			name: "no match",
			args: map[string]any{"text": "hello", "pattern": "\\d+"},
			want: []string{},
		},
		{
			name:    "invalid regex",
			args:    map[string]any{"text": "hello", "pattern": "("},
			wantErr: "invalid regex:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, TextExtractRegex, tc.args)
			if tc.wantErr != "" {
				assert.True(t, result.IsError)
				assert.Contains(t, resultText(t, result), tc.wantErr)
				return
			}

			assert.False(t, result.IsError)
			var got []string
			require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &got))
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTextBase64Encode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		text    string
		want    string
		isError bool
	}{
		{name: "basic encode", text: "hello", want: base64.StdEncoding.EncodeToString([]byte("hello"))},
		{name: "empty string", text: "", want: "text is required", isError: true},
		{name: "unicode", text: "你好", want: base64.StdEncoding.EncodeToString([]byte("你好"))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, TextBase64Encode, map[string]any{"text": tc.text})
			assert.Equal(t, tc.isError, result.IsError)
			assert.Equal(t, tc.want, resultText(t, result))
		})
	}
}

func TestTextBase64Decode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		encoded string
		want    string
		isError bool
	}{
		{name: "valid decode", encoded: base64.StdEncoding.EncodeToString([]byte("relay")), want: "relay"},
		{name: "invalid base64", encoded: "%%%not-base64%%%", want: "decode base64:", isError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, TextBase64Decode, map[string]any{"encoded": tc.encoded})
			assert.Equal(t, tc.isError, result.IsError)
			if tc.isError {
				assert.Contains(t, resultText(t, result), tc.want)
				return
			}
			assert.Equal(t, tc.want, resultText(t, result))
		})
	}
}

func TestTextMarkdownToHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		markdown string
		want     []string
		isError  bool
	}{
		{name: "headings", markdown: "# Title", want: []string{"<h1>Title</h1>"}},
		{name: "bold", markdown: "**bold**", want: []string{"<p><strong>bold</strong></p>"}},
		{name: "italic", markdown: "*italic*", want: []string{"<p><em>italic</em></p>"}},
		{name: "links", markdown: "[site](https://example.com)", want: []string{`<a href="https://example.com">site</a>`}},
		{name: "code blocks", markdown: "```go\nfmt.Println(\"hi\")\n```", want: []string{"<pre><code>fmt.Println(&#34;hi&#34;)</code></pre>"}},
		{
			name:     "mixed",
			markdown: "# Title\n\nThis is **bold** and *italic* with `code` and [link](https://example.com).",
			want: []string{
				"<h1>Title</h1>",
				"<strong>bold</strong>",
				"<em>italic</em>",
				"<code>code</code>",
				`<a href="https://example.com">link</a>`,
			},
		},
		{name: "empty markdown", markdown: "", want: []string{"markdown is required"}, isError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, TextMarkdownToHTML, map[string]any{"markdown": tc.markdown})
			assert.Equal(t, tc.isError, result.IsError)
			body := resultText(t, result)
			for _, want := range tc.want {
				assert.Contains(t, body, want)
			}
		})
	}
}
