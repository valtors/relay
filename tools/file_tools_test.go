package tools

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRead(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "sample.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello file"), 0o644))

	tests := []struct {
		name    string
		path    string
		want    string
		isError bool
	}{
		{name: "read existing file", path: filePath, want: "hello file"},
		{name: "read nonexistent file", path: filepath.Join(dir, "missing.txt"), want: "read file:", isError: true},
		{name: "read directory", path: dir, want: "read file:", isError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := callTool(t, FileRead, map[string]any{"path": tc.path})
			assert.Equal(t, tc.isError, result.IsError)
			if tc.isError {
				assert.Contains(t, resultText(t, result), tc.want)
				return
			}
			assert.Equal(t, tc.want, resultText(t, result))
		})
	}
}

func TestFileWrite(t *testing.T) {
	t.Parallel()

	t.Run("write new file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "new.txt")
		result := callTool(t, FileWrite, map[string]any{"path": path, "content": "new content"})

		assert.False(t, result.IsError)
		assert.Equal(t, path, resultText(t, result))

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "new content", string(data))
	})

	t.Run("overwrite existing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "overwrite.txt")
		require.NoError(t, os.WriteFile(path, []byte("old"), 0o644))

		result := callTool(t, FileWrite, map[string]any{"path": path, "content": "updated"})
		assert.False(t, result.IsError)

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "updated", string(data))
	})

	t.Run("create nested dirs", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a", "b", "c.txt")

		result := callTool(t, FileWrite, map[string]any{"path": path, "content": "nested"})
		assert.False(t, result.IsError)

		data, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "nested", string(data))
	})

	t.Run("empty content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")

		result := callTool(t, FileWrite, map[string]any{"path": path, "content": ""})
		assert.False(t, result.IsError)

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.EqualValues(t, 0, info.Size())
	})
}

func TestFileList(t *testing.T) {
	t.Parallel()

	t.Run("list flat dir", func(t *testing.T) {
		dir := t.TempDir()
		alpha := filepath.Join(dir, "alpha.txt")
		beta := filepath.Join(dir, "beta")
		require.NoError(t, os.WriteFile(alpha, []byte("a"), 0o644))
		require.NoError(t, os.Mkdir(beta, 0o755))

		result := callTool(t, FileList, map[string]any{"path": dir})
		assert.False(t, result.IsError)
		assert.Equal(t, []string{alpha, beta}, strings.Split(resultText(t, result), "\n"))
	})

	t.Run("list recursive", func(t *testing.T) {
		dir := t.TempDir()
		nestedDir := filepath.Join(dir, "nested")
		inner := filepath.Join(nestedDir, "inner.txt")
		rootFile := filepath.Join(dir, "root.txt")
		require.NoError(t, os.MkdirAll(nestedDir, 0o755))
		require.NoError(t, os.WriteFile(rootFile, []byte("root"), 0o644))
		require.NoError(t, os.WriteFile(inner, []byte("inner"), 0o644))

		result := callTool(t, FileList, map[string]any{"path": dir, "recursive": true})
		assert.False(t, result.IsError)
		assert.Equal(t, []string{nestedDir, inner, rootFile}, strings.Split(resultText(t, result), "\n"))
	})

	t.Run("empty dir", func(t *testing.T) {
		dir := t.TempDir()
		result := callTool(t, FileList, map[string]any{"path": dir})
		assert.False(t, result.IsError)
		assert.Equal(t, "", resultText(t, result))
	})

	t.Run("nonexistent dir", func(t *testing.T) {
		dir := t.TempDir()
		missing := filepath.Join(dir, "missing")
		result := callTool(t, FileList, map[string]any{"path": missing})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "stat path:")
	})
}

func TestFileSize(t *testing.T) {
	t.Parallel()

	t.Run("normal file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "kib.txt")
		require.NoError(t, os.WriteFile(path, []byte(strings.Repeat("a", 1024)), 0o644))

		result := callTool(t, FileSize, map[string]any{"path": path})
		assert.False(t, result.IsError)
		assert.Equal(t, "1.0 KiB", resultText(t, result))
	})

	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")
		require.NoError(t, os.WriteFile(path, nil, 0o644))

		result := callTool(t, FileSize, map[string]any{"path": path})
		assert.False(t, result.IsError)
		assert.Equal(t, "0 B", resultText(t, result))
	})

	t.Run("nonexistent file", func(t *testing.T) {
		dir := t.TempDir()
		result := callTool(t, FileSize, map[string]any{"path": filepath.Join(dir, "missing.txt")})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "stat file:")
	})
}

func TestFileHash(t *testing.T) {
	t.Parallel()

	t.Run("known hash value", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "hello.txt")
		require.NoError(t, os.WriteFile(path, []byte("hello"), 0o644))

		result := callTool(t, FileHash, map[string]any{"path": path})
		assert.False(t, result.IsError)
		assert.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte("hello"))), resultText(t, result))
	})

	t.Run("nonexistent file", func(t *testing.T) {
		dir := t.TempDir()
		result := callTool(t, FileHash, map[string]any{"path": filepath.Join(dir, "missing.txt")})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "read file:")
	})
}

func TestFileZipAndUnzipRoundTrip(t *testing.T) {
	t.Parallel()

	sourceRoot := t.TempDir()
	sourceDir := filepath.Join(sourceRoot, "payload")
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "root.txt"), []byte("root-data"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "nested", "child.txt"), []byte("child-data"), 0o644))

	archivePath := filepath.Join(t.TempDir(), "archive.zip")
	zipResult := callTool(t, FileZip, map[string]any{
		"paths":  []any{sourceDir},
		"output": archivePath,
	})
	assert.False(t, zipResult.IsError)
	assert.Equal(t, archivePath, resultText(t, zipResult))

	unzipDir := t.TempDir()
	unzipResult := callTool(t, FileUnzip, map[string]any{
		"path":       archivePath,
		"output_dir": unzipDir,
	})
	assert.False(t, unzipResult.IsError)
	assert.Equal(t, unzipDir, resultText(t, unzipResult))

	rootData, err := os.ReadFile(filepath.Join(unzipDir, "payload", "root.txt"))
	require.NoError(t, err)
	assert.Equal(t, "root-data", string(rootData))

	childData, err := os.ReadFile(filepath.Join(unzipDir, "payload", "nested", "child.txt"))
	require.NoError(t, err)
	assert.Equal(t, "child-data", string(childData))
}
