package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func WebFetch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := strings.TrimSpace(req.GetString("url", ""))
	if url == "" {
		return mcp.NewToolResultError("url is required"), nil
	}

	method := strings.TrimSpace(req.GetString("method", ""))
	if method == "" {
		method = http.MethodGet
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("build request: %v", err)), nil
	}

	for key, value := range readHeaderArgs(req.GetArguments()["headers"]) {
		for _, item := range value {
			httpReq.Header.Add(key, item)
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("fetch url: %v", err)), nil
	}
	defer resp.Body.Close()

	body, truncated, err := readResponseCapped(resp.Body, maxToolFileBytes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("read response: %v", err)), nil
	}
	if truncated {
		return mcp.NewToolResultText(string(body) + "\n\n[truncated at 1MB]"), nil
	}
	return mcp.NewToolResultText(string(body)), nil
}

func WebStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := strings.TrimSpace(req.GetString("url", ""))
	if url == "" {
		return mcp.NewToolResultError("url is required"), nil
	}

	client := &http.Client{Timeout: 15 * time.Second}
	start := time.Now()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("build request: %v", err)), nil
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("check url: %v", err)), nil
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	return mcp.NewToolResultText(fmt.Sprintf("%d (%s)", resp.StatusCode, time.Since(start).Round(time.Millisecond))), nil
}

func readHeaderArgs(raw any) map[string][]string {
	result := make(map[string][]string)
	headers, ok := raw.(map[string]any)
	if !ok {
		return result
	}
	for key, value := range headers {
		switch v := value.(type) {
		case string:
			result[key] = append(result[key], v)
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok {
					result[key] = append(result[key], s)
				}
			}
		}
	}
	return result
}

func readResponseCapped(reader io.Reader, maxBytes int64) ([]byte, bool, error) {
	var buf bytes.Buffer
	read, err := io.CopyN(&buf, reader, maxBytes+1)
	if err != nil && err != io.EOF {
		return nil, false, err
	}
	data := buf.Bytes()
	if read > maxBytes {
		return data[:maxBytes], true, nil
	}
	return data, false, nil
}
