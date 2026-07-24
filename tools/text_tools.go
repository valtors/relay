package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func TextWordCount(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text := req.GetString("text", "")
	if text == "" {
		return mcp.NewToolResultError("text is required"), nil
	}

	count := len(strings.Fields(text))
	return mcp.NewToolResultText(fmt.Sprintf("%d", count)), nil
}

func TextReplace(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text := req.GetString("text", "")
	find := req.GetString("find", "")
	replace := req.GetString("replace", "")
	if text == "" {
		return mcp.NewToolResultError("text is required"), nil
	}
	if find == "" {
		return mcp.NewToolResultError("find is required"), nil
	}

	if req.GetBool("all", true) {
		return mcp.NewToolResultText(strings.ReplaceAll(text, find, replace)), nil
	}
	return mcp.NewToolResultText(strings.Replace(text, find, replace, 1)), nil
}

func TextExtractRegex(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text := req.GetString("text", "")
	pattern := req.GetString("pattern", "")
	if text == "" {
		return mcp.NewToolResultError("text is required"), nil
	}
	if pattern == "" {
		return mcp.NewToolResultError("pattern is required"), nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid regex: %v", err)), nil
	}

	matches := re.FindAllString(text, -1)
	if matches == nil {
		matches = []string{}
	}
	output, err := json.Marshal(matches)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("encode matches: %v", err)), nil
	}
	return mcp.NewToolResultText(string(output)), nil
}

func TextBase64Encode(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text := req.GetString("text", "")
	if text == "" {
		return mcp.NewToolResultError("text is required"), nil
	}
	return mcp.NewToolResultText(base64.StdEncoding.EncodeToString([]byte(text))), nil
}

func TextBase64Decode(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	encoded := req.GetString("encoded", "")
	if encoded == "" {
		return mcp.NewToolResultError("encoded is required"), nil
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("decode base64: %v", err)), nil
	}
	return mcp.NewToolResultText(string(decoded)), nil
}

func TextMarkdownToHTML(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := req.GetString("markdown", "")
	if input == "" {
		return mcp.NewToolResultError("markdown is required"), nil
	}
	return mcp.NewToolResultText(renderMarkdownHTML(input)), nil
}

var (
	linkPattern   = regexp.MustCompile(`\[(.+?)\]\((.+?)\)`)
	boldPattern   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicPattern = regexp.MustCompile(`\*(.+?)\*`)
	codePattern   = regexp.MustCompile("`(.+?)`")
)

func renderMarkdownHTML(input string) string {
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	var output bytes.Buffer
	var paragraph []string
	inCodeBlock := false
	var codeLines []string

	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}
		output.WriteString("<p>")
		output.WriteString(applyInlineMarkdown(strings.Join(paragraph, " ")))
		output.WriteString("</p>\n")
		paragraph = nil
	}

	flushCodeBlock := func() {
		output.WriteString("<pre><code>")
		output.WriteString(html.EscapeString(strings.Join(codeLines, "\n")))
		output.WriteString("</code></pre>\n")
		codeLines = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			flushParagraph()
			if inCodeBlock {
				flushCodeBlock()
			}
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}
		if trimmed == "" {
			flushParagraph()
			continue
		}
		if level := headingLevel(trimmed); level > 0 {
			flushParagraph()
			content := strings.TrimSpace(trimmed[level:])
			fmt.Fprintf(&output, "<h%d>%s</h%d>\n", level, applyInlineMarkdown(content), level)
			continue
		}
		paragraph = append(paragraph, trimmed)
	}

	flushParagraph()
	if inCodeBlock {
		flushCodeBlock()
	}

	return strings.TrimSpace(output.String())
}

func applyInlineMarkdown(text string) string {
	escaped := html.EscapeString(text)
	escaped = linkPattern.ReplaceAllStringFunc(escaped, func(match string) string {
		parts := linkPattern.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		href := sanitizeURL(parts[2])
		return fmt.Sprintf(`<a href="%s">%s</a>`, href, parts[1])
	})
	escaped = boldPattern.ReplaceAllString(escaped, `<strong>$1</strong>`)
	escaped = codePattern.ReplaceAllString(escaped, `<code>$1</code>`)
	escaped = italicPattern.ReplaceAllString(escaped, `<em>$1</em>`)
	return escaped
}

func sanitizeURL(raw string) string {
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		decoded = raw
	}
	trimmed := strings.TrimSpace(decoded)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") || strings.HasPrefix(lower, "vbscript:") {
		return "#"
	}
	return raw
}

func headingLevel(line string) int {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return 0
	}
	if len(line) == level || line[level] != ' ' {
		return 0
	}
	return level
}
