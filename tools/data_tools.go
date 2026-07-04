package tools

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func DataJSONFormat(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := req.GetString("json", "")
	if input == "" {
		return mcp.NewToolResultError("json is required"), nil
	}

	var value any
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	formatted, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(string(formatted)), nil
}

func DataCSVToJSON(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := req.GetString("csv", "")
	if input == "" {
		return mcp.NewToolResultError("csv is required"), nil
	}

	reader := csv.NewReader(strings.NewReader(input))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parse csv: %v", err)), nil
	}
	if len(rows) == 0 {
		return mcp.NewToolResultText("[]"), nil
	}

	headers := normalizeHeaders(rows[0])
	records := make([]map[string]string, 0, maxInt(len(rows)-1, 0))
	for _, row := range rows[1:] {
		if len(row) > len(headers) {
			for i := len(headers); i < len(row); i++ {
				headers = append(headers, fmt.Sprintf("column_%d", i+1))
			}
		}
		record := make(map[string]string, len(headers))
		for i, header := range headers {
			if i < len(row) {
				record[header] = row[i]
				continue
			}
			record[header] = ""
		}
		records = append(records, record)
	}

	output, err := json.Marshal(records)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("encode json: %v", err)), nil
	}
	return mcp.NewToolResultText(string(output)), nil
}

func DataJSONToCSV(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := req.GetString("json", "")
	if input == "" {
		return mcp.NewToolResultError("json is required"), nil
	}

	var items []map[string]any
	if err := json.Unmarshal([]byte(input), &items); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parse json: %v", err)), nil
	}

	headers := make([]string, 0)
	seen := make(map[string]struct{})
	for _, item := range items {
		for key := range item {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			headers = append(headers, key)
		}
	}

	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("write csv: %v", err)), nil
		}
	}
	for _, item := range items {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = csvValue(item[header])
		}
		if err := writer.Write(row); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("write csv: %v", err)), nil
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("flush csv: %v", err)), nil
	}

	return mcp.NewToolResultText(builder.String()), nil
}

func DataJSONQuery(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := req.GetString("json", "")
	if input == "" {
		return mcp.NewToolResultError("json is required"), nil
	}

	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	var value any
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("parse json: %v", err)), nil
	}

	current := value
	for _, segment := range strings.Split(path, ".") {
		switch node := current.(type) {
		case map[string]any:
			next, ok := node[segment]
			if !ok {
				return mcp.NewToolResultError(fmt.Sprintf("path not found: %s", segment)), nil
			}
			current = next
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid array index: %s", segment)), nil
			}
			if index < 0 || index >= len(node) {
				return mcp.NewToolResultError(fmt.Sprintf("array index out of range: %d", index)), nil
			}
			current = node[index]
		default:
			return mcp.NewToolResultError(fmt.Sprintf("path not found: %s", segment)), nil
		}
	}

	return mcp.NewToolResultText(formatJSONQueryValue(current)), nil
}

func normalizeHeaders(headers []string) []string {
	out := make([]string, len(headers))
	seen := make(map[string]int)
	for i, header := range headers {
		name := strings.TrimSpace(header)
		if name == "" {
			name = fmt.Sprintf("column_%d", i+1)
		}
		seen[name]++
		if seen[name] > 1 {
			name = fmt.Sprintf("%s_%d", name, seen[name])
		}
		out[i] = name
	}
	return out
}

func csvValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

func formatJSONQueryValue(value any) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
