// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Tamish Max
package state

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestWriteAndReadOutput(t *testing.T) {
	filename := "test_output_" + t.Name() + ".md"
	WriteOutput(filename, "test content")
	data, err := ReadOutput(filename)
	if err != nil {
		t.Fatalf("ReadOutput: %v", err)
	}
	if data != "test content" {
		t.Errorf("expected 'test content', got %s", data)
	}
}

func TestReadOutput_NoFile(t *testing.T) {
	_, err := ReadOutput("nonexistent_file_12345.md")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestOutputExists(t *testing.T) {
	filename := "exists_unique_" + t.Name() + ".md"
	os.Remove(filepath.Join(mustOutputDir(t), filename))
	if OutputExists(filename) {
		t.Error("expected false")
	}
	WriteOutput(filename, "x")
	if !OutputExists(filename) {
		t.Error("expected true after writing")
	}
	os.Remove(filepath.Join(mustOutputDir(t), filename))
}

func TestAcquireAndReleaseLock(t *testing.T) {
	lockName := "test_lock_" + t.Name()
	if err := AcquireLock(lockName); err != nil {
		t.Fatalf("AcquireLock: %v", err)
	}
	ReleaseLock(lockName)
	if err := AcquireLock(lockName); err != nil {
		t.Errorf("AcquireLock after release: %v", err)
	}
	ReleaseLock(lockName)
}

func TestReleaseLock_NoLock(t *testing.T) {
	ReleaseLock("nonexistent_lock_12345")
}

func TestReadBrief(t *testing.T) {
	dir := t.TempDir()
	briefPath := filepath.Join(dir, "brief.md")
	brief := "this is a brief"
	os.WriteFile(briefPath, []byte(brief), 0644)
	data, err := ReadBrief(briefPath)
	if err != nil {
		t.Fatalf("ReadBrief: %v", err)
	}
	if data != brief {
		t.Errorf("expected %s, got %s", brief, data)
	}
}

func TestReadBrief_NoFile(t *testing.T) {
	_, err := ReadBrief(filepath.Join(t.TempDir(), "nonexistent.md"))
	if err == nil {
		t.Error("expected error for missing brief")
	}
}

func TestOutputDir(t *testing.T) {
	dir, err := OutputDir()
	if err != nil {
		t.Fatalf("OutputDir: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty dir")
	}
}

func TestInitSession(t *testing.T) {
	dir := t.TempDir()
	briefPath := filepath.Join(dir, "brief.md")
	os.WriteFile(briefPath, []byte("test brief"), 0644)
	err := InitSession(briefPath)
	if err != nil {
		t.Fatalf("InitSession: %v", err)
	}
}

func mustOutputDir(t *testing.T) string {
	t.Helper()
	dir, err := OutputDir()
	if err != nil {
		t.Fatalf("OutputDir: %v", err)
	}
	return dir
}

func TestIsStageComplete_NotFound(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("RELAY_SESSION_DIR", dir)
	defer os.Unsetenv("RELAY_SESSION_DIR")
	result := IsStageComplete("nonexistent_stage")
	if result {
		t.Error("expected false for nonexistent stage")
	}
}

func TestIncrementIteration(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("RELAY_SESSION_DIR", dir)
	defer os.Unsetenv("RELAY_SESSION_DIR")
	n, err := IncrementIteration("test_stage")
	if err != nil {
		t.Fatalf("IncrementIteration: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
	n2, _ := IncrementIteration("test_stage")
	if n2 != 2 {
		t.Errorf("expected 2, got %d", n2)
	}
}

func TestAcquireLock_DoubleAcquire(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("RELAY_SESSION_DIR", dir)
	defer os.Unsetenv("RELAY_SESSION_DIR")
	err := AcquireLock("test_double")
	if err != nil {
		t.Fatalf("first AcquireLock: %v", err)
	}
	err2 := AcquireLock("test_double")
	if err2 == nil {
		t.Error("expected error on second acquire")
	}
	ReleaseLock("test_double")
}

func TestAcquireLock_StaleLock(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	filename := "test.txt"
	lp, _ := lockPath(filename)
	os.MkdirAll(filepath.Dir(lp), 0755)
	os.WriteFile(lp, []byte("999999"), 0o644)

	err := AcquireLock(filename)
	if err != nil {
		t.Fatalf("should clear stale lock: %v", err)
	}

	b, _ := os.ReadFile(lp)
	pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	if pid != os.Getpid() {
		t.Errorf("expected own pid, got %d", pid)
	}
	ReleaseLock(filename)
}

func TestAcquireLock_OwnPid(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "test2.txt")

	lp, _ := lockPath(filename)
	os.WriteFile(lp, []byte(strconv.Itoa(os.Getpid())), 0o644)

	err := AcquireLock(filename)
	if err == nil {
		t.Error("should error when locked by own pid")
		ReleaseLock(filename)
	}
}

func TestWriteOutput_NestedDir(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	err := WriteOutput("test.txt", "hello world")
	if err != nil {
		t.Errorf("WriteOutput failed: %v", err)
	}

	od, _ := OutputDir()
	data, _ := os.ReadFile(filepath.Join(od, "test.txt"))
	if string(data) != "hello world" {
		t.Errorf("content mismatch: %s", data)
	}
}

func TestReadBrief_NonExistent(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	_, err := ReadBrief(filepath.Join(dir, "nonexistent.md"))
	if err == nil {
		t.Error("should error on nonexistent file")
	}
}
