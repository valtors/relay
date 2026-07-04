package tools

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

type pdfPageDimension struct {
	Page   int     `json:"page"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func PDFInfoTool(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	file, err := os.Open(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("open pdf: %v", err)), nil
	}
	defer file.Close()

	info, err := api.PDFInfo(file, filepath.Base(resolved), nil, false, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("read pdf info: %v", err)), nil
	}

	dims, err := api.PageDimsFile(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("read page dimensions: %v", err)), nil
	}

	dimensions := make([]pdfPageDimension, 0, len(dims))
	for i, dim := range dims {
		dimensions = append(dimensions, pdfPageDimension{
			Page:   i + 1,
			Width:  dim.Width,
			Height: dim.Height,
		})
	}

	return pdfJSONResult(map[string]any{
		"pages":      info.PageCount,
		"title":      info.Title,
		"author":     info.Author,
		"creator":    info.Creator,
		"dimensions": dimensions,
	}), nil
}

func PDFExtractTextTool(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	selectedPages, err := parsePDFPageSelection(req.GetString("pages", ""), false)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ctx, err := api.ReadContextFile(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("read pdf: %v", err)), nil
	}

	pages, err := api.PagesForPageSelection(ctx.PageCount, selectedPages, true, true)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parse pages: %v", err)), nil
	}

	parts := make([]string, 0, ctx.PageCount)
	for page := 1; page <= ctx.PageCount; page++ {
		if !pages[page] {
			continue
		}

		reader, err := pdfcpu.ExtractPageContent(ctx, page)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("extract page %d content: %v", page, err)), nil
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("read page %d content: %v", page, err)), nil
		}

		text := strings.TrimSpace(extractPDFTextContent(data))
		if text != "" {
			parts = append(parts, text)
		}
	}

	return mcp.NewToolResultText(strings.TrimSpace(strings.Join(parts, "\n\n"))), nil
}

func PDFPageCountTool(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	count, err := api.PageCountFile(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("count pdf pages: %v", err)), nil
	}

	return mcp.NewToolResultText(strconv.Itoa(count)), nil
}

func PDFMergeTool(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	paths, err := req.RequireStringSlice("paths")
	if err != nil || len(paths) == 0 {
		return mcp.NewToolResultError("paths is required"), nil
	}
	if len(paths) < 2 {
		return mcp.NewToolResultError("paths must contain at least 2 PDFs"), nil
	}

	resolvedPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		resolved, err := resolveToolPath(path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		resolvedPaths = append(resolvedPaths, resolved)
	}

	output := strings.TrimSpace(req.GetString("output", ""))
	if output == "" {
		return mcp.NewToolResultError("output is required"), nil
	}

	outputPath, err := resolveToolPath(output)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create output directory: %v", err)), nil
	}

	if err := api.MergeCreateFile(resolvedPaths, outputPath, false, nil); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("merge pdfs: %v", err)), nil
	}

	return mcp.NewToolResultText(outputPath), nil
}

func PDFSplitTool(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	outputDirRaw := strings.TrimSpace(req.GetString("output_dir", ""))
	outputDir := defaultPDFSplitOutputDir(resolved)
	if outputDirRaw != "" {
		outputDir, err = resolveToolPath(outputDirRaw)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	before, err := listPDFsInDir(outputDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list output directory: %v", err)), nil
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create output directory: %v", err)), nil
	}

	if err := api.SplitFile(resolved, outputDir, 1, nil); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("split pdf: %v", err)), nil
	}

	after, err := listPDFsInDir(outputDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list split output files: %v", err)), nil
	}

	newFiles := make([]string, 0, len(after))
	for path := range after {
		if _, ok := before[path]; !ok {
			newFiles = append(newFiles, path)
		}
	}
	if len(newFiles) == 0 {
		for path := range after {
			newFiles = append(newFiles, path)
		}
	}
	sort.Strings(newFiles)

	return pdfJSONResult(newFiles), nil
}

func PDFExtractPagesTool(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	selectedPages, err := parsePDFPageSelection(req.GetString("pages", ""), true)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	output := strings.TrimSpace(req.GetString("output", ""))
	if output == "" {
		return mcp.NewToolResultError("output is required"), nil
	}

	outputPath, err := resolveToolPath(output)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create output directory: %v", err)), nil
	}

	if err := api.TrimFile(resolved, outputPath, selectedPages, nil); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("extract pdf pages: %v", err)), nil
	}

	return mcp.NewToolResultText(outputPath), nil
}

func parsePDFPageSelection(raw string, required bool) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return nil, fmt.Errorf("pages is required")
		}
		return nil, nil
	}

	selectedPages, err := api.ParsePageSelection(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid pages value: %w", err)
	}
	return selectedPages, nil
}

func defaultPDFSplitOutputDir(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return filepath.Join(filepath.Dir(path), base+"_split")
}

func listPDFsInDir(dir string) (map[string]struct{}, error) {
	entries := map[string]struct{}{}

	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("output_dir must be a directory")
	}

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range dirEntries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".pdf") {
			continue
		}
		entries[filepath.Join(dir, entry.Name())] = struct{}{}
	}

	return entries, nil
}

func extractPDFTextContent(content []byte) string {
	var lines []string
	var current strings.Builder
	var lastText string
	var lastArrayText string

	flushLine := func() {
		line := strings.TrimSpace(current.String())
		if line != "" {
			lines = append(lines, line)
		}
		current.Reset()
	}

	appendText := func(text string, newLine bool) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		if newLine && current.Len() > 0 {
			flushLine()
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(text)
	}

	for i := 0; i < len(content); {
		switch content[i] {
		case ' ', '\t', '\r', '\n', '\f', 0:
			i++
		case '%':
			for i < len(content) && content[i] != '\n' && content[i] != '\r' {
				i++
			}
		case '(':
			text, next, err := parsePDFLiteralString(content, i)
			if err != nil {
				return strings.TrimSpace(string(content))
			}
			lastText = text
			i = next
		case '[':
			text, next, err := parsePDFArrayText(content, i)
			if err != nil {
				return strings.TrimSpace(string(content))
			}
			lastArrayText = text
			i = next
		case '<':
			if i+1 < len(content) && content[i+1] == '<' {
				i += 2
				continue
			}
			text, next, err := parsePDFHexString(content, i)
			if err != nil {
				return strings.TrimSpace(string(content))
			}
			lastText = text
			i = next
		case '\'', '"':
			appendText(lastText, true)
			lastText = ""
			lastArrayText = ""
			i++
		default:
			token, next := parsePDFOperator(content, i)
			switch token {
			case "Tj":
				appendText(lastText, false)
				lastText = ""
			case "TJ":
				appendText(lastArrayText, false)
				lastArrayText = ""
			case "Td", "TD", "T*", "ET":
				flushLine()
			}
			i = next
		}
	}

	flushLine()
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func parsePDFLiteralString(content []byte, start int) (string, int, error) {
	var out bytes.Buffer
	depth := 0

	for i := start; i < len(content); i++ {
		ch := content[i]
		if i == start {
			depth = 1
			continue
		}

		if ch == '\\' {
			if i+1 >= len(content) {
				return out.String(), len(content), nil
			}
			next := content[i+1]
			switch next {
			case 'n':
				out.WriteByte('\n')
			case 'r':
				out.WriteByte('\r')
			case 't':
				out.WriteByte('\t')
			case 'b':
				out.WriteByte('\b')
			case 'f':
				out.WriteByte('\f')
			case '\\', '(', ')':
				out.WriteByte(next)
			case '\n':
			case '\r':
				if i+2 < len(content) && content[i+2] == '\n' {
					i++
				}
			default:
				if next >= '0' && next <= '7' {
					end := i + 2
					for end < len(content) && end < i+4 && content[end] >= '0' && content[end] <= '7' {
						end++
					}
					value, err := strconv.ParseInt(string(content[i+1:end]), 8, 32)
					if err != nil {
						return "", 0, err
					}
					out.WriteByte(byte(value))
					i = end - 1
					continue
				}
				out.WriteByte(next)
			}
			i++
			continue
		}

		if ch == '(' {
			depth++
			out.WriteByte(ch)
			continue
		}
		if ch == ')' {
			depth--
			if depth == 0 {
				return out.String(), i + 1, nil
			}
			out.WriteByte(ch)
			continue
		}

		out.WriteByte(ch)
	}

	return "", 0, fmt.Errorf("unterminated PDF string")
}

func parsePDFHexString(content []byte, start int) (string, int, error) {
	end := start + 1
	for end < len(content) && content[end] != '>' {
		end++
	}
	if end >= len(content) {
		return "", 0, fmt.Errorf("unterminated PDF hex string")
	}

	raw := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\r', '\n', '\f':
			return -1
		default:
			return r
		}
	}, string(content[start+1:end]))
	if len(raw)%2 == 1 {
		raw += "0"
	}

	decoded, err := hex.DecodeString(raw)
	if err != nil {
		return "", 0, err
	}
	return string(decoded), end + 1, nil
}

func parsePDFArrayText(content []byte, start int) (string, int, error) {
	var parts []string
	depth := 0

	for i := start; i < len(content); {
		ch := content[i]
		if ch == '[' {
			depth++
			i++
			continue
		}
		if ch == ']' {
			depth--
			i++
			if depth == 0 {
				return strings.Join(parts, ""), i, nil
			}
			continue
		}
		if ch == '(' {
			text, next, err := parsePDFLiteralString(content, i)
			if err != nil {
				return "", 0, err
			}
			parts = append(parts, text)
			i = next
			continue
		}
		if ch == '<' && !(i+1 < len(content) && content[i+1] == '<') {
			text, next, err := parsePDFHexString(content, i)
			if err != nil {
				return "", 0, err
			}
			parts = append(parts, text)
			i = next
			continue
		}
		i++
	}

	return "", 0, fmt.Errorf("unterminated PDF text array")
}

func parsePDFOperator(content []byte, start int) (string, int) {
	end := start
	for end < len(content) {
		switch content[end] {
		case ' ', '\t', '\r', '\n', '\f', 0, '[', ']', '(', ')', '<', '>', '%':
			return string(content[start:end]), end
		default:
			end++
		}
	}
	return string(content[start:end]), end
}

func pdfJSONResult(v any) *mcp.CallToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err))
	}
	return mcp.NewToolResultText(string(data))
}
