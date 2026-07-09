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

func tempWorkDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })
	return dir
}

func TestFileRead(t *testing.T) {
	dir := tempWorkDir(t)
	filePath := filepath.Join(dir, "sample.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello file"), 0o644))

	tests := []struct {
		name    string
		path    string
		want    string
		isError bool
	}{
		{name: "read existing file", path: "sample.txt", want: "hello file"},
		{name: "read nonexistent file", path: "missing.txt", want: "read file:", isError: true},
		{name: "read directory", path: ".", want: "read file:", isError: true},
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
	t.Run("write new file", func(t *testing.T) {
		dir := tempWorkDir(t)
		result := callTool(t, FileWrite, map[string]any{"path": "new.txt", "content": "new content"})
		assert.False(t, result.IsError)
		assert.Equal(t, filepath.Join(dir, "new.txt"), resultText(t, result))

		data, err := os.ReadFile("new.txt")
		require.NoError(t, err)
		assert.Equal(t, "new content", string(data))
	})

	t.Run("overwrite existing", func(t *testing.T) {
		tempWorkDir(t)
		require.NoError(t, os.WriteFile("overwrite.txt", []byte("old"), 0o644))

		result := callTool(t, FileWrite, map[string]any{"path": "overwrite.txt", "content": "updated"})
		assert.False(t, result.IsError)

		data, err := os.ReadFile("overwrite.txt")
		require.NoError(t, err)
		assert.Equal(t, "updated", string(data))
	})

	t.Run("create nested dirs", func(t *testing.T) {
		tempWorkDir(t)
		result := callTool(t, FileWrite, map[string]any{"path": "a/b/c.txt", "content": "nested"})
		assert.False(t, result.IsError)

		data, err := os.ReadFile("a/b/c.txt")
		require.NoError(t, err)
		assert.Equal(t, "nested", string(data))
	})

	t.Run("empty content", func(t *testing.T) {
		tempWorkDir(t)
		result := callTool(t, FileWrite, map[string]any{"path": "empty.txt", "content": ""})
		assert.False(t, result.IsError)

		info, err := os.Stat("empty.txt")
		require.NoError(t, err)
		assert.EqualValues(t, 0, info.Size())
	})
}

func TestFileList(t *testing.T) {
	t.Run("list flat dir", func(t *testing.T) {
		dir := tempWorkDir(t)
		alpha := filepath.Join(dir, "alpha.txt")
		beta := filepath.Join(dir, "beta")
		require.NoError(t, os.WriteFile(alpha, []byte("a"), 0o644))
		require.NoError(t, os.Mkdir(beta, 0o755))

		result := callTool(t, FileList, map[string]any{"path": "."})
		assert.False(t, result.IsError)
		assert.Equal(t, []string{alpha, beta}, strings.Split(resultText(t, result), "\n"))
	})

	t.Run("list recursive", func(t *testing.T) {
		dir := tempWorkDir(t)
		nestedDir := filepath.Join(dir, "nested")
		inner := filepath.Join(nestedDir, "inner.txt")
		rootFile := filepath.Join(dir, "root.txt")
		require.NoError(t, os.MkdirAll(nestedDir, 0o755))
		require.NoError(t, os.WriteFile(rootFile, []byte("root"), 0o644))
		require.NoError(t, os.WriteFile(inner, []byte("inner"), 0o644))

		result := callTool(t, FileList, map[string]any{"path": ".", "recursive": true})
		assert.False(t, result.IsError)
		assert.Equal(t, []string{nestedDir, inner, rootFile}, strings.Split(resultText(t, result), "\n"))
	})

	t.Run("empty dir", func(t *testing.T) {
		tempWorkDir(t)
		subdir := "empty_sub"
		require.NoError(t, os.Mkdir(subdir, 0o755))
		result := callTool(t, FileList, map[string]any{"path": subdir})
		assert.False(t, result.IsError)
		assert.Equal(t, "", resultText(t, result))
	})

	t.Run("nonexistent dir", func(t *testing.T) {
		tempWorkDir(t)
		result := callTool(t, FileList, map[string]any{"path": "missing"})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "stat path:")
	})
}

func TestFileSize(t *testing.T) {
	t.Run("normal file", func(t *testing.T) {
		tempWorkDir(t)
		require.NoError(t, os.WriteFile("kib.txt", []byte(strings.Repeat("a", 1024)), 0o644))

		result := callTool(t, FileSize, map[string]any{"path": "kib.txt"})
		assert.False(t, result.IsError)
		assert.Equal(t, "1.0 KiB", resultText(t, result))
	})

	t.Run("empty file", func(t *testing.T) {
		tempWorkDir(t)
		require.NoError(t, os.WriteFile("empty.txt", nil, 0o644))

		result := callTool(t, FileSize, map[string]any{"path": "empty.txt"})
		assert.False(t, result.IsError)
		assert.Equal(t, "0 B", resultText(t, result))
	})

	t.Run("nonexistent file", func(t *testing.T) {
		tempWorkDir(t)
		result := callTool(t, FileSize, map[string]any{"path": "missing.txt"})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "stat file:")
	})
}

func TestFileHash(t *testing.T) {
	t.Run("known hash value", func(t *testing.T) {
		tempWorkDir(t)
		require.NoError(t, os.WriteFile("hello.txt", []byte("hello"), 0o644))

		result := callTool(t, FileHash, map[string]any{"path": "hello.txt"})
		assert.False(t, result.IsError)
		assert.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte("hello"))), resultText(t, result))
	})

	t.Run("nonexistent file", func(t *testing.T) {
		tempWorkDir(t)
		result := callTool(t, FileHash, map[string]any{"path": "missing.txt"})
		assert.True(t, result.IsError)
		assert.Contains(t, resultText(t, result), "read file:")
	})
}

func TestFileZipAndUnzipRoundTrip(t *testing.T) {
	dir := tempWorkDir(t)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "payload", "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "payload", "root.txt"), []byte("root-data"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "payload", "nested", "child.txt"), []byte("child-data"), 0o644))

	zipResult := callTool(t, FileZip, map[string]any{
		"paths":  []any{"payload"},
		"output": "archive.zip",
	})
	assert.False(t, zipResult.IsError)
	assert.Equal(t, filepath.Join(dir, "archive.zip"), resultText(t, zipResult))

	unzipResult := callTool(t, FileUnzip, map[string]any{
		"path":       "archive.zip",
		"output_dir": "extracted",
	})
	assert.False(t, unzipResult.IsError)

	rootData, err := os.ReadFile(filepath.Join(dir, "extracted", "payload", "root.txt"))
	require.NoError(t, err)
	assert.Equal(t, "root-data", string(rootData))

	childData, err := os.ReadFile(filepath.Join(dir, "extracted", "payload", "nested", "child.txt"))
	require.NoError(t, err)
	assert.Equal(t, "child-data", string(childData))
}

func TestFileRead_PathTraversal(t *testing.T) {
	tempWorkDir(t)
	result := callTool(t, FileRead, map[string]any{"path": "../../../etc/passwd"})
	assert.True(t, result.IsError)
	assert.Contains(t, resultText(t, result), "outside working directory")
}
