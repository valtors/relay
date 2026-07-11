package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	neturl "net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const defaultWebTimeout = 30 * time.Second

func webTimeout() time.Duration {
	raw := os.Getenv("RELAY_WEB_TIMEOUT")
	if raw == "" {
		return defaultWebTimeout
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 1 {
		return defaultWebTimeout
	}
	return time.Duration(seconds) * time.Second
}

func WebFetch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := strings.TrimSpace(req.GetString("url", ""))
	if url == "" {
		return mcp.NewToolResultError("url is required"), nil
	}

	if err := validateURL(url); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
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

	client := &http.Client{Timeout: webTimeout()}
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

	if err := validateURL(url); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client := &http.Client{Timeout: webTimeout()}
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

func validateURL(raw string) error {
	if os.Getenv("RELAY_SKIP_URL_VALIDATION") != "" {
		return nil
	}

	parsed, err := neturl.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("url must have a hostname")
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("url resolves to a non-public address: %s", host)
		}
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("dns lookup failed for %s: %w", host, err)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("url resolves to a non-public address: %s -> %s", host, ip)
		}
	}

	blocked := []string{"169.254.169.254", "metadata.google.internal", "metadata.aws.internal"}
	for _, b := range blocked {
		if strings.EqualFold(host, b) {
			return fmt.Errorf("url points to a cloud metadata endpoint: %s", host)
		}
	}

	return nil
}
