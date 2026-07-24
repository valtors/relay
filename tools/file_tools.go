package tools

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

const maxToolFileBytes = 1 << 20

func FileHash(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := req.GetString("path", "")
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("read file: %v", err)), nil
	}

	sum := sha256.Sum256(data)
	return mcp.NewToolResultText(fmt.Sprintf("%x", sum)), nil
}

func FileRead(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, truncated, err := readFileCapped(resolved, maxToolFileBytes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("read file: %v", err)), nil
	}
	if truncated {
		return mcp.NewToolResultText(string(data) + "\n\n[truncated at 1MB]"), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func FileWrite(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	args := req.GetArguments()
	rawContent, ok := args["content"]
	if !ok {
		return mcp.NewToolResultError("content is required"), nil
	}
	content, ok := rawContent.(string)
	if !ok {
		return mcp.NewToolResultError("content must be a string"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create directories: %v", err)), nil
	}
	if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("write file: %v", err)), nil
	}

	return mcp.NewToolResultText(resolved), nil
}

func FileList(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("stat path: %v", err)), nil
	}
	if !info.IsDir() {
		return mcp.NewToolResultError("path must be a directory"), nil
	}

	recursive := req.GetBool("recursive", false)
	entries := make([]string, 0)
	if recursive {
		err = filepath.WalkDir(resolved, func(current string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if current == resolved {
				return nil
			}
			entries = append(entries, current)
			return nil
		})
	} else {
		dirEntries, readErr := os.ReadDir(resolved)
		if readErr != nil {
			err = readErr
		} else {
			for _, entry := range dirEntries {
				entries = append(entries, filepath.Join(resolved, entry.Name()))
			}
		}
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list directory: %v", err)), nil
	}

	sort.Strings(entries)
	return mcp.NewToolResultText(strings.Join(entries, "\n")), nil
}

func FileSize(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	resolved, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("stat file: %v", err)), nil
	}

	return mcp.NewToolResultText(humanReadableSize(info.Size())), nil
}

func FileZip(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	paths, err := req.RequireStringSlice("paths")
	if err != nil || len(paths) == 0 {
		return mcp.NewToolResultError("paths is required"), nil
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

	outFile, err := os.Create(outputPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("create archive: %v", err)), nil
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)

	for _, rawPath := range paths {
		sourcePath, pathErr := resolveToolPath(rawPath)
		if pathErr != nil {
			return mcp.NewToolResultError(pathErr.Error()), nil
		}
		if addErr := addPathToZip(zw, sourcePath); addErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("zip %s: %v", sourcePath, addErr)), nil
		}
	}

	if err := zw.Close(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("finalize archive: %v", err)), nil
	}

	return mcp.NewToolResultText(outputPath), nil
}

func FileUnzip(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := strings.TrimSpace(req.GetString("path", ""))
	if path == "" {
		return mcp.NewToolResultError("path is required"), nil
	}

	archivePath, err := resolveToolPath(path)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	outputDir := strings.TrimSpace(req.GetString("output_dir", ""))
	if outputDir == "" {
		outputDir, err = os.Getwd()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get current directory: %v", err)), nil
		}
	} else {
		outputDir, err = resolveToolPath(outputDir)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	if err := unzipArchive(archivePath, outputDir); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("extract archive: %v", err)), nil
	}

	return mcp.NewToolResultText(outputDir), nil
}

func resolveToolPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is required")
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current directory: %w", err)
	}
	wd = filepath.Clean(wd)

	var resolved string
	if filepath.IsAbs(path) {
		resolved = filepath.Clean(path)
	} else {
		resolved = filepath.Clean(filepath.Join(wd, path))
	}

	rel, err := filepath.Rel(wd, resolved)
	if err != nil {
		return "", fmt.Errorf("path outside working directory: %s", path)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path outside working directory: %s", path)
	}

	return resolved, nil
}

func readFileCapped(path string, maxBytes int64) ([]byte, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer file.Close()

	var buf bytes.Buffer
	read, err := io.CopyN(&buf, file, maxBytes+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, false, err
	}
	data := buf.Bytes()
	if read > maxBytes {
		return data[:maxBytes], true, nil
	}
	return data, false, nil
}

func humanReadableSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func addPathToZip(zw *zip.Writer, sourcePath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	baseName := filepath.Base(sourcePath)
	if info.IsDir() {
		return filepath.Walk(sourcePath, func(path string, _ os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			currentInfo, statErr := os.Stat(path)
			if statErr != nil {
				return statErr
			}
			rel, relErr := filepath.Rel(sourcePath, path)
			if relErr != nil {
				return relErr
			}
			name := filepath.ToSlash(filepath.Join(baseName, rel))
			if rel == "." {
				name = filepath.ToSlash(baseName) + "/"
			}
			return writeZipEntry(zw, path, currentInfo, name)
		})
	}

	return writeZipEntry(zw, sourcePath, info, filepath.ToSlash(baseName))
}

func writeZipEntry(zw *zip.Writer, path string, info os.FileInfo, name string) error {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = name
	if info.IsDir() {
		if !strings.HasSuffix(header.Name, "/") {
			header.Name += "/"
		}
		_, err = zw.CreateHeader(header)
		return err
	}
	header.Method = zip.Deflate
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	return err
}

func unzipArchive(archivePath, outputDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	cleanOutput := filepath.Clean(outputDir)
	if err := os.MkdirAll(cleanOutput, 0o755); err != nil {
		return err
	}

	for _, file := range reader.File {
		targetPath := filepath.Join(cleanOutput, file.Name)
		cleanTarget := filepath.Clean(targetPath)
		if cleanTarget != cleanOutput && !strings.HasPrefix(cleanTarget, cleanOutput+string(os.PathSeparator)) {
			return fmt.Errorf("invalid zip path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanTarget, 0o755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanTarget), 0o755); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}

		dst, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = src.Close()
			return err
		}

		_, copyErr := io.Copy(dst, src)
		closeErr := dst.Close()
		srcErr := src.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if srcErr != nil {
			return srcErr
		}
	}

	return nil
}
