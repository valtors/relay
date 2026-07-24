package search

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Result struct {
	Title   string
	URL     string
	Snippet string
}

var (
	httpClient = &http.Client{Timeout: 20 * time.Second}

	reTitle   = regexp.MustCompile(`(?s)class="result__a"\s+href="([^"]+)"[^>]*>(.*?)</a>`)
	reSnippet = regexp.MustCompile(`(?s)class="result__snippet"[^>]*>(.*?)</a>`)
	reTags    = regexp.MustCompile(`<[^>]+>`)
	reSpaces  = regexp.MustCompile(`\s+`)
)

func DDG(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 5
	}
	form := url.Values{"q": {query}}
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://html.duckduckgo.com/html/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ddg request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ddg read: %w", err)
	}
	html := string(body)

	titles := reTitle.FindAllStringSubmatch(html, -1)
	snippets := reSnippet.FindAllStringSubmatch(html, -1)

	out := make([]Result, 0, limit)
	for i, t := range titles {
		if i >= limit {
			break
		}
		r := Result{
			URL:   cleanURL(t[1]),
			Title: cleanText(t[2]),
		}
		if i < len(snippets) {
			r.Snippet = cleanText(snippets[i][1])
		}
		if r.URL == "" || r.Title == "" {
			continue
		}
		out = append(out, r)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("ddg: no results parsed (got %d bytes)", len(html))
	}
	return out, nil
}

func DDGMulti(ctx context.Context, queries []string, perQuery int) []Result {
	seen := map[string]bool{}
	out := []Result{}
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		results, err := DDG(ctx, q, perQuery)
		if err != nil {
			continue
		}
		for _, r := range results {
			if seen[r.URL] {
				continue
			}
			seen[r.URL] = true
			out = append(out, r)
		}
	}
	return out
}

func FormatForPrompt(results []Result) string {
	if len(results) == 0 {
		return "(no search results)"
	}
	var sb strings.Builder
	for i, r := range results {
		title := sanitizeForPrompt(r.Title)
		snippet := sanitizeForPrompt(r.Snippet)
		fmt.Fprintf(&sb, "<result index=\"%d\">\n  <title>%s</title>\n  <url>%s</url>\n  <snippet>%s</snippet>\n</result>\n", i+1, title, r.URL, snippet)
	}
	return sb.String()
}

func sanitizeForPrompt(s string) string {
	s = strings.ReplaceAll(s, "--- WEB_SEARCH_RESULTS ---", "")
	s = strings.ReplaceAll(s, "--- END WEB_SEARCH_RESULTS ---", "")
	s = strings.ReplaceAll(s, "</result>", "")
	s = strings.ReplaceAll(s, "<result", "")
	return s
}

func cleanURL(u string) string {
	if strings.HasPrefix(u, "//duckduckgo.com/l/") || strings.HasPrefix(u, "/l/") {
		if idx := strings.Index(u, "uddg="); idx >= 0 {
			rest := u[idx+5:]
			if amp := strings.Index(rest, "&"); amp >= 0 {
				rest = rest[:amp]
			}
			if dec, err := url.QueryUnescape(rest); err == nil {
				return dec
			}
		}
	}
	return u
}

func cleanText(s string) string {
	s = reTags.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&#x27;", "'")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = reSpaces.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
