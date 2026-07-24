package search

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withHTTPClient(t *testing.T, client *http.Client) {
	t.Helper()
	original := httpClient
	httpClient = client
	t.Cleanup(func() {
		httpClient = original
	})
}

func TestFormatForPrompt_Empty(t *testing.T) {
	result := FormatForPrompt([]Result{})
	assert.Equal(t, "(no search results)", result)
}

func TestFormatForPrompt_Nil(t *testing.T) {
	result := FormatForPrompt(nil)
	assert.Equal(t, "(no search results)", result)
}

func TestFormatForPrompt_Multiple(t *testing.T) {
	results := []Result{
		{Title: "First", URL: "https://example.com/1", Snippet: "Snippet one"},
		{Title: "Second", URL: "https://example.com/2", Snippet: "Snippet two"},
	}
	result := FormatForPrompt(results)
	assert.Contains(t, result, "First")
	assert.Contains(t, result, "https://example.com/1")
	assert.Contains(t, result, "Snippet one")
	assert.Contains(t, result, "Second")
	assert.Contains(t, result, "https://example.com/2")
	assert.Contains(t, result, "Snippet two")
}

func TestFormatForPrompt_Single(t *testing.T) {
	results := []Result{
		{Title: "Only", URL: "https://only.com", Snippet: "only snippet"},
	}
	result := FormatForPrompt(results)
	assert.Contains(t, result, "Only")
	assert.Contains(t, result, "https://only.com")
	assert.Contains(t, result, "only snippet")
}

func TestCleanURL_DirectURL(t *testing.T) {
	assert.Equal(t, "https://example.com", cleanURL("https://example.com"))
}

func TestCleanURL_DuckDuckGoRedirect(t *testing.T) {
	raw := "//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com%2Fpath&rut=abc"
	assert.Equal(t, "https://example.com/path", cleanURL(raw))
}

func TestCleanURL_DuckDuckGoRedirectShort(t *testing.T) {
	raw := "/l/?uddg=https%3A%2F%2Fexample.org"
	assert.Equal(t, "https://example.org", cleanURL(raw))
}

func TestCleanURL_NoUDDGParam(t *testing.T) {
	raw := "//duckduckgo.com/l/?rut=abc"
	assert.Equal(t, raw, cleanURL(raw))
}

func TestCleanURL_NotRedirect(t *testing.T) {
	assert.Equal(t, "https://plain.com/page", cleanURL("https://plain.com/page"))
}

func TestCleanText_StripsTags(t *testing.T) {
	assert.Equal(t, "hello world", cleanText("<b>hello</b> <i>world</i>"))
}

func TestCleanText_DecodesEntities(t *testing.T) {
	assert.Equal(t, `a & b "c" <d> 'e'`, cleanText("a &amp; b &quot;c&quot; &lt;d&gt; &#x27;e&#x27;"))
}

func TestCleanText_DecodesAposEntity(t *testing.T) {
	assert.Equal(t, "it's", cleanText("it&#39;s"))
}

func TestCleanText_DecodesNbsp(t *testing.T) {
	assert.Equal(t, "a b", cleanText("a&nbsp;b"))
}

func TestCleanText_CollapsesWhitespace(t *testing.T) {
	assert.Equal(t, "a b c", cleanText("a   b\n\n\tc"))
}

func TestCleanText_Empty(t *testing.T) {
	assert.Equal(t, "", cleanText(""))
}

func TestCleanText_OnlyTags(t *testing.T) {
	assert.Equal(t, "", cleanText("<b></b>"))
}

func TestParseRegex_TitleMatches(t *testing.T) {
	html := `<a class="result__a" href="https://example.com">Title Here</a>`
	matches := reTitle.FindAllStringSubmatch(html, -1)
	require.Len(t, matches, 1)
	assert.Equal(t, "https://example.com", matches[0][1])
	assert.Equal(t, "Title Here", matches[0][2])
}

func TestParseRegex_SnippetMatches(t *testing.T) {
	html := `<a class="result__snippet" href="#">Snippet text</a>`
	matches := reSnippet.FindAllStringSubmatch(html, -1)
	require.Len(t, matches, 1)
	assert.Equal(t, "Snippet text", matches[0][1])
}

func TestParseRegex_MultipleResults(t *testing.T) {
	html := `
	<a class="result__a" href="https://a.com/1">A1</a>
	<a class="result__snippet">s1</a>
	<a class="result__a" href="https://a.com/2">A2</a>
	<a class="result__snippet">s2</a>
	<a class="result__a" href="https://a.com/3">A3</a>
	<a class="result__snippet">s3</a>
	`
	titles := reTitle.FindAllStringSubmatch(html, -1)
	snippets := reSnippet.FindAllStringSubmatch(html, -1)
	assert.Len(t, titles, 3)
	assert.Len(t, snippets, 3)
	assert.Equal(t, "https://a.com/1", titles[0][1])
	assert.Equal(t, "A1", cleanText(titles[0][2]))
	assert.Equal(t, "s3", cleanText(snippets[2][1]))
}

func TestParseRegex_NoMatches(t *testing.T) {
	html := `<html><body>nothing here</body></html>`
	titles := reTitle.FindAllStringSubmatch(html, -1)
	snippets := reSnippet.FindAllStringSubmatch(html, -1)
	assert.Nil(t, titles)
	assert.Nil(t, snippets)
}

func TestDDG_BuildsQueryAndParsesResults(t *testing.T) {
	withHTTPClient(t, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "https://html.duckduckgo.com/html/", req.URL.String())
		assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		assert.NotEmpty(t, req.Header.Get("User-Agent"))

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		form, err := url.ParseQuery(string(body))
		require.NoError(t, err)
		assert.Equal(t, "relay search", form.Get("q"))

		html := `
			<a class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com%2Frelay&amp;rut=abc">Relay &amp; Search</a>
			<a class="result__snippet">First <b>snippet</b></a>
			<a class="result__a" href="https://example.org/direct">Direct Result</a>
			<a class="result__snippet">Second&nbsp;snippet</a>
		`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(html)),
			Header:     make(http.Header),
		}, nil
	})})

	results, err := DDG(context.Background(), "relay search", 2)
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, Result{
		Title:   "Relay & Search",
		URL:     "https://example.com/relay",
		Snippet: "First snippet",
	}, results[0])
	assert.Equal(t, Result{
		Title:   "Direct Result",
		URL:     "https://example.org/direct",
		Snippet: "Second snippet",
	}, results[1])
}

func TestDDG_MalformedResponseReturnsParseError(t *testing.T) {
	withHTTPClient(t, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`<html><body>no search result anchors</body></html>`)),
			Header:     make(http.Header),
		}, nil
	})})

	results, err := DDG(context.Background(), "relay search", 5)
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "ddg: no results parsed")
}

func TestDDG_NetworkError(t *testing.T) {
	withHTTPClient(t, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network unavailable")
	})})

	results, err := DDG(context.Background(), "relay search", 5)
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "ddg request")
	assert.Contains(t, err.Error(), "network unavailable")
}

func TestDDG_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := DDG(ctx, "test", 5)
	assert.Error(t, err)
}

func TestDDGMulti_SkipsEmptyQueries(t *testing.T) {
	withHTTPClient(t, &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("DDGMulti should not issue requests for empty queries")
		return nil, nil //nolint:nilnil // intentional: test assertion path
	})})

	results := DDGMulti(context.Background(), []string{"", "  ", "\t\n"}, 5)
	assert.Empty(t, results)
}

func TestDDGMulti_AllEmpty(t *testing.T) {
	results := DDGMulti(context.Background(), []string{"", "  "}, 5)
	assert.Len(t, results, 0)
}

func TestDDGMulti_EmptyInput(t *testing.T) {
	results := DDGMulti(context.Background(), nil, 5)
	assert.Len(t, results, 0)
}

func TestResult_Fields(t *testing.T) {
	r := Result{Title: "T", URL: "U", Snippet: "S"}
	assert.Equal(t, "T", r.Title)
	assert.Equal(t, "U", r.URL)
	assert.Equal(t, "S", r.Snippet)
}

func TestCleanURL_LultipleParams(t *testing.T) {
	raw := "//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com%2Fpage%3Ffoo%3Dbar&rut=xyz&other=1"
	assert.Equal(t, "https://example.com/page?foo=bar", cleanURL(raw))
}

func TestReTags(t *testing.T) {
	re := regexp.MustCompile(`<[^>]+>`)
	assert.Equal(t, "hello", re.ReplaceAllString("<b>hello</b>", ""))
}

func TestReSpaces(t *testing.T) {
	re := regexp.MustCompile(`\s+`)
	assert.Equal(t, "a b c", re.ReplaceAllString("a   b\n\n\tc", " "))
}

func TestFormatForPrompt_ExactFormatting(t *testing.T) {
	results := []Result{
		{Title: "Test", URL: "https://test.com", Snippet: "test snippet"},
	}
	result := FormatForPrompt(results)
	assert.Contains(t, result, "<title>Test</title>")
	assert.Contains(t, result, "<url>https://test.com</url>")
	assert.Contains(t, result, "<snippet>test snippet</snippet>")
}

func TestCleanText_NestedTags(t *testing.T) {
	assert.Equal(t, "inner text", cleanText("<div><span>inner text</span></div>"))
}

func TestCleanText_MixedEntitiesAndTags(t *testing.T) {
	input := `<b>Hello</b> &amp; <i>world</i>&nbsp;<a>link</a>`
	assert.Equal(t, "Hello & world link", cleanText(input))
}

func TestCleanURL_DuckDuckGoPrefixWithoutUDDG(t *testing.T) {
	raw := "//duckduckgo.com/l/?uddg="
	assert.Equal(t, "", cleanURL(raw))
}

func TestFormatForPrompt_LongSnippet(t *testing.T) {
	longSnippet := strings.Repeat("x", 500)
	results := []Result{
		{Title: "T", URL: "U", Snippet: longSnippet},
	}
	result := FormatForPrompt(results)
	assert.Contains(t, result, longSnippet)
}
