package tools

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataJSONFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		isError bool
	}{
		{
			name:  "valid json",
			input: `{"b":2,"a":1}`,
			want:  "{\n  \"a\": 1,\n  \"b\": 2\n}",
		},
		{
			name:  "nested json",
			input: `{"user":{"name":"Ada","tags":["go","mcp"]}}`,
			want:  "{\n  \"user\": {\n    \"name\": \"Ada\",\n    \"tags\": [\n      \"go\",\n      \"mcp\"\n    ]\n  }\n}",
		},
		{
			name:    "invalid json",
			input:   `{"a":`,
			want:    "unexpected end of JSON input",
			isError: true,
		},
		{
			name:  "already formatted",
			input: "{\n  \"a\": 1\n}",
			want:  "{\n  \"a\": 1\n}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, DataJSONFormat, map[string]any{"json": tc.input})
			assert.Equal(t, tc.isError, result.IsError)
			assert.Equal(t, tc.want, resultText(t, result))
		})
	}
}

func TestDataCSVToJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		want      []map[string]string
		wantError string
	}{
		{
			name:  "basic csv",
			input: "name,age\nAda,37\nBob,41",
			want: []map[string]string{
				{"name": "Ada", "age": "37"},
				{"name": "Bob", "age": "41"},
			},
		},
		{
			name:  "headers with spaces",
			input: " First Name , Last Name \nAda,Lovelace",
			want: []map[string]string{
				{"First Name": "Ada", "Last Name": "Lovelace"},
			},
		},
		{
			name:      "empty csv",
			input:     "",
			wantError: "csv is required",
		},
		{
			name:  "single row",
			input: "name,role\nAda,engineer",
			want: []map[string]string{
				{"name": "Ada", "role": "engineer"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, DataCSVToJSON, map[string]any{"csv": tc.input})
			if tc.wantError != "" {
				assert.True(t, result.IsError)
				assert.Equal(t, tc.wantError, resultText(t, result))
				return
			}

			assert.False(t, result.IsError)
			var got []map[string]string
			require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &got))
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDataJSONToCSV(t *testing.T) {
	t.Parallel()

	t.Run("basic array of objects", func(t *testing.T) {
		result := callTool(t, DataJSONToCSV, map[string]any{
			"json": `[{"name":"Ada","age":37},{"name":"Bob","age":41}]`,
		})
		assert.False(t, result.IsError)

		rows := parseCSVResult(t, resultText(t, result))
		require.Len(t, rows, 3)
		assert.ElementsMatch(t, []string{"name", "age"}, rows[0])
		assertCSVRowMatches(t, rows[0], rows[1], map[string]string{"name": "Ada", "age": "37"})
		assertCSVRowMatches(t, rows[0], rows[2], map[string]string{"name": "Bob", "age": "41"})
	})

	t.Run("nested objects are serialized", func(t *testing.T) {
		result := callTool(t, DataJSONToCSV, map[string]any{
			"json": `[{"name":"Ada","meta":{"city":"London"}}]`,
		})
		assert.False(t, result.IsError)

		rows := parseCSVResult(t, resultText(t, result))
		require.Len(t, rows, 2)
		assert.ElementsMatch(t, []string{"name", "meta"}, rows[0])
		assertCSVRowMatches(t, rows[0], rows[1], map[string]string{"name": "Ada", "meta": `{"city":"London"}`})
	})

	t.Run("empty array", func(t *testing.T) {
		result := callTool(t, DataJSONToCSV, map[string]any{"json": `[]`})
		assert.False(t, result.IsError)
		assert.Equal(t, "", resultText(t, result))
	})
}

func TestDataJSONQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    map[string]any
		want    string
		isError bool
	}{
		{
			name: "simple path",
			args: map[string]any{"json": `{"name":"Ada"}`, "path": "name"},
			want: "Ada",
		},
		{
			name: "nested path",
			args: map[string]any{"json": `{"user":{"name":"Ada"}}`, "path": "user.name"},
			want: "Ada",
		},
		{
			name: "array index",
			args: map[string]any{"json": `{"items":["a","b","c"]}`, "path": "items.1"},
			want: "b",
		},
		{
			name:    "missing path",
			args:    map[string]any{"json": `{"name":"Ada"}`, "path": "user.name"},
			want:    "path not found:",
			isError: true,
		},
		{
			name:    "invalid json",
			args:    map[string]any{"json": `{"name":`, "path": "name"},
			want:    "parse json:",
			isError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, DataJSONQuery, tc.args)
			assert.Equal(t, tc.isError, result.IsError)
			if tc.isError {
				assert.Contains(t, resultText(t, result), tc.want)
				return
			}
			assert.Equal(t, tc.want, resultText(t, result))
		})
	}
}

func parseCSVResult(t *testing.T, input string) [][]string {
	t.Helper()

	reader := csv.NewReader(strings.NewReader(input))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	require.NoError(t, err)
	return rows
}

func assertCSVRowMatches(t *testing.T, headers, row []string, want map[string]string) {
	t.Helper()

	got := make(map[string]string, len(headers))
	for i, header := range headers {
		if i < len(row) {
			got[header] = row[i]
		}
	}
	assert.Equal(t, want, got)
}
