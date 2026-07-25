package state

import (
	"os"
	"path/filepath"
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
