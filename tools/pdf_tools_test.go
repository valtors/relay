package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPdfPageCount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "three-pages.pdf")
	writeTestPDF(t, path, testPDFSpec{Pages: 3})

	result := callTool(t, PDFPageCountTool, map[string]any{"path": path})
	require.False(t, result.IsError)
	assert.Equal(t, "3", resultText(t, result))
}

func TestPdfInfo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "info.pdf")
	writeTestPDF(t, path, testPDFSpec{
		Pages:   2,
		Title:   "Quarterly Report",
		Author:  "Relay",
		Creator: "pdf_tools_test",
	})

	result := callTool(t, PDFInfoTool, map[string]any{"path": path})
	require.False(t, result.IsError)

	var payload struct {
		Pages      int `json:"pages"`
		Title      string
		Author     string
		Creator    string
		Dimensions []struct {
			Page   int
			Width  float64
			Height float64
		} `json:"dimensions"`
	}
	require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &payload))

	assert.Equal(t, 2, payload.Pages)
	assert.Equal(t, "Quarterly Report", payload.Title)
	assert.Equal(t, "Relay", payload.Author)
	assert.Equal(t, "pdf_tools_test", payload.Creator)
	require.Len(t, payload.Dimensions, 2)
	assert.Equal(t, 1, payload.Dimensions[0].Page)
	assert.Equal(t, 300.0, payload.Dimensions[0].Width)
	assert.Equal(t, 400.0, payload.Dimensions[0].Height)
}

func TestPdfExtractText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "text.pdf")
	writeTestPDF(t, path, testPDFSpec{
		Pages: 2,
		PageText: []string{
			"Hello page one",
			"Hello page two",
		},
	})

	result := callTool(t, PDFExtractTextTool, map[string]any{
		"path":  path,
		"pages": "2",
	})
	require.False(t, result.IsError)
	assert.Equal(t, "Hello page two", strings.TrimSpace(resultText(t, result)))
}

func TestPdfMerge(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first.pdf")
	second := filepath.Join(dir, "second.pdf")
	output := filepath.Join(dir, "merged.pdf")
	writeTestPDF(t, first, testPDFSpec{Pages: 1})
	writeTestPDF(t, second, testPDFSpec{Pages: 2})

	result := callTool(t, PDFMergeTool, map[string]any{
		"paths":  []string{first, second},
		"output": output,
	})
	require.False(t, result.IsError)
	assert.Equal(t, output, resultText(t, result))

	count, err := api.PageCountFile(output)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestPdfSplit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "split.pdf")
	outputDir := filepath.Join(dir, "parts")
	writeTestPDF(t, path, testPDFSpec{Pages: 3})

	result := callTool(t, PDFSplitTool, map[string]any{
		"path":       path,
		"output_dir": outputDir,
	})
	require.False(t, result.IsError)

	var files []string
	require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &files))
	require.Len(t, files, 3)
	for _, file := range files {
		assert.FileExists(t, file)
	}
}

func TestPdfExtractPages(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "source.pdf")
	output := filepath.Join(dir, "extracted.pdf")
	writeTestPDF(t, path, testPDFSpec{Pages: 3})

	result := callTool(t, PDFExtractPagesTool, map[string]any{
		"path":   path,
		"pages":  "1,3",
		"output": output,
	})
	require.False(t, result.IsError)
	assert.Equal(t, output, resultText(t, result))

	count, err := api.PageCountFile(output)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestPdfErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing.pdf")
		result := callTool(t, PDFPageCountTool, map[string]any{"path": path})
		require.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "count pdf pages")
	})

	t.Run("invalid pdf", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "invalid.pdf")
		require.NoError(t, os.WriteFile(path, []byte("not a real pdf"), 0o644))

		result := callTool(t, PDFInfoTool, map[string]any{"path": path})
		require.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "read pdf info")
	})
}

type testPDFSpec struct {
	Pages    int
	Title    string
	Author   string
	Creator  string
	PageText []string
}

func writeTestPDF(t *testing.T, path string, spec testPDFSpec) {
	t.Helper()

	if spec.Pages <= 0 {
		spec.Pages = 1
	}
	if len(spec.PageText) < spec.Pages {
		for i := len(spec.PageText); i < spec.Pages; i++ {
			spec.PageText = append(spec.PageText, fmt.Sprintf("Page %d", i+1))
		}
	}

	pageObjNums := make([]int, spec.Pages)
	contentObjNums := make([]int, spec.Pages)
	nextObjNum := 5
	for i := 0; i < spec.Pages; i++ {
		pageObjNums[i] = nextObjNum
		nextObjNum++
		contentObjNums[i] = nextObjNum
		nextObjNum++
	}

	objects := map[int]string{
		1: "<< /Type /Catalog /Pages 2 0 R >>",
		2: fmt.Sprintf("<< /Type /Pages /Count %d /Kids [%s] >>", spec.Pages, joinObjectRefs(pageObjNums)),
		3: "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		4: fmt.Sprintf("<< /Title (%s) /Author (%s) /Creator (%s) >>",
			escapePDFString(spec.Title),
			escapePDFString(spec.Author),
			escapePDFString(spec.Creator),
		),
	}

	for i := 0; i < spec.Pages; i++ {
		stream := fmt.Sprintf("BT /F1 12 Tf 72 300 Td (%s) Tj ET", escapePDFString(spec.PageText[i]))
		objects[pageObjNums[i]] = fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 400] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>",
			contentObjNums[i],
		)
		objects[contentObjNums[i]] = fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream)
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")

	offsets := make([]int, nextObjNum)
	for i := 1; i < nextObjNum; i++ {
		offsets[i] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i, objects[i])
	}

	xrefOffset := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", nextObjNum)
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i < nextObjNum; i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R /Info 4 0 R >>\nstartxref\n%d\n%%%%EOF\n", nextObjNum, xrefOffset)

	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o644))
}

func joinObjectRefs(objNums []int) string {
	refs := make([]string, 0, len(objNums))
	for _, objNum := range objNums {
		refs = append(refs, fmt.Sprintf("%d 0 R", objNum))
	}
	return strings.Join(refs, " ")
}

func escapePDFString(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"(", "\\(",
		")", "\\)",
		"\r", "\\r",
		"\n", "\\n",
	)
	return replacer.Replace(s)
}
