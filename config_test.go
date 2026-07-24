package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintConfigUsage(t *testing.T) {
	var buf bytes.Buffer
	printConfigUsage(&buf)
	out := buf.String()
	if !strings.Contains(out, "relay config") {
		t.Fatalf("printConfigUsage = %q, want usage with 'relay config'", out)
	}
}

func TestMinFunc(t *testing.T) {
	if got := min(3, 7); got != 3 {
		t.Fatalf("min(3,7) = %d, want 3", got)
	}
	if got := min(10, 4); got != 4 {
		t.Fatalf("min(10,4) = %d, want 4", got)
	}
	if got := min(5, 5); got != 5 {
		t.Fatalf("min(5,5) = %d, want 5", got)
	}
}

func TestRunConfigCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ui := cliUI{color: false}
	code := runConfigCommand([]string{}, &stdout, &stderr, ui)
	if code != 0 {
		t.Fatalf("runConfigCommand code = %d, want 0", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "relay config") {
		t.Fatalf("output missing 'relay config': %q", out)
	}
	if !strings.Contains(out, "tools:") {
		t.Fatalf("output missing 'tools:' line: %q", out)
	}
	if !strings.Contains(out, "environment") {
		t.Fatalf("output missing 'environment' section: %q", out)
	}
	if !strings.Contains(out, "ANTHROPIC_API_KEY") {
		t.Fatalf("output missing ANTHROPIC_API_KEY: %q", out)
	}
}

func TestRunConfigCommand_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ui := cliUI{color: false}
	code := runConfigCommand([]string{"--help"}, &stdout, &stderr, ui)
	if code != 0 {
		t.Fatalf("runConfigCommand --help code = %d, want 0", code)
	}
}

func TestRunConfigCommand_InvalidArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ui := cliUI{color: false}
	code := runConfigCommand([]string{"bogus"}, &stdout, &stderr, ui)
	if code != 1 {
		t.Fatalf("runConfigCommand bogus code = %d, want 1", code)
	}
}
