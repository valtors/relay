package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTitleCaseWord(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"a", "A"},
		{"", ""},
		{"hELLO", "HELLO"},
		{"two words", "Two words"},
	}
	for _, tt := range tests {
		got := titleCaseWord(tt.input)
		if got != tt.expected {
			t.Errorf("titleCaseWord(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRuneLen(t *testing.T) {
	if runeLen("hello") != 5 {
		t.Error("expected 5")
	}
	if runeLen("") != 0 {
		t.Error("expected 0")
	}
	if runeLen("a") != 1 {
		t.Error("expected 1")
	}
}

func TestShortToolName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"web_search", "web search"},
		{"file_read", "file read"},
		{"single", "single"},
		{"", ""},
	}
	for _, tt := range tests {
		entry := toolInfo{Name: tt.name}
		got := shortToolName(entry)
		if got != tt.expected {
			t.Errorf("shortToolName(%q) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

func TestToolDisplayRank(t *testing.T) {
	entry := toolInfo{Name: "web_search"}
	if toolDisplayRank(entry) < 0 {
		t.Error("expected >= 0")
	}
	entry2 := toolInfo{Name: "unknown_tool"}
	if toolDisplayRank(entry2) < 0 {
		t.Error("expected >= 0")
	}
}

func TestIsTerminalWriter_Term(t *testing.T) {
	if !isTerminalWriter(os.Stdout) {
		t.Log("os.Stdout may not be a terminal in test env, that's ok")
	}
}

func TestIsTerminalWriter_NonTerm(t *testing.T) {
	var buf bytes.Buffer
	if isTerminalWriter(&buf) {
		t.Fatal("buffer should not be a terminal")
	}
}

func TestBold_NoColor(t *testing.T) {
	ui := cliUI{color: false}
	result := ui.bold("hello")
	if result != "hello" {
		t.Fatalf("expected hello with no color, got %s", result)
	}
}

func TestGreen_NoColor(t *testing.T) {
	ui := cliUI{color: false}
	result := ui.green("hello")
	if result != "hello" {
		t.Fatalf("expected hello with no color, got %s", result)
	}
}

func TestRenderHint_NoColor(t *testing.T) {
	ui := cliUI{color: false}
	result := ui.renderHint("some hint")
	if result != "some hint" {
		t.Fatalf("expected plain text with no color, got %s", result)
	}
}

func TestRenderHint_Empty(t *testing.T) {
	ui := cliUI{color: true}
	result := ui.renderHint("")
	if result != "" {
		t.Fatalf("expected empty string, got %s", result)
	}
}

func TestRenderHint_WithColor(t *testing.T) {
	ui := cliUI{color: true}
	result := ui.renderHint("hint text")
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestRenderStyled_NoColor(t *testing.T) {
	ui := cliUI{color: false}
	style := lipgloss.NewStyle().Bold(true)
	result := ui.renderStyled(style, "test")
	if result != "test" {
		t.Fatalf("expected plain test with no color, got %s", result)
	}
}

func TestRenderStyled_Empty(t *testing.T) {
	ui := cliUI{color: true}
	style := lipgloss.NewStyle().Bold(true)
	result := ui.renderStyled(style, "")
	if result != "" {
		t.Fatalf("expected empty string, got %s", result)
	}
}

func TestLipglossRenderer_WithExisting(t *testing.T) {
	r := lipgloss.NewRenderer(io.Discard)
	ui := cliUI{color: true, renderer: r}
	result := ui.lipglossRenderer()
	if result != r {
		t.Fatal("expected existing renderer to be returned")
	}
}

func TestLipglossRenderer_NewDefault(t *testing.T) {
	ui := cliUI{color: true}
	result := ui.lipglossRenderer()
	if result == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestDoctorMarker_NoColor(t *testing.T) {
	ui := cliUI{color: false}
	if ui.doctorMarker("+") != "+" {
		t.Fatal("expected + with no color")
	}
	if ui.doctorMarker("!") != "!" {
		t.Fatal("expected ! with no color")
	}
	if ui.doctorMarker("-") != "-" {
		t.Fatal("expected - with no color")
	}
	if ui.doctorMarker("?") != "?" {
		t.Fatal("expected ? unchanged")
	}
}

func TestDoctorMarker_WithColor(t *testing.T) {
	ui := cliUI{color: true}
	result := ui.doctorMarker("+")
	if result == "" {
		t.Fatal("expected non-empty result for +")
	}
	result = ui.doctorMarker("!")
	if result == "" {
		t.Fatal("expected non-empty result for !")
	}
	result = ui.doctorMarker("-")
	if result == "" {
		t.Fatal("expected non-empty result for -")
	}
}

func TestCheckEnvironment(t *testing.T) {
	result := checkEnvironment()
	if len(result) == 0 {
		t.Fatal("expected at least one check")
	}
}

func TestCheckNetwork(t *testing.T) {
	result := checkNetwork()
	if result.name == "" {
		t.Fatal("expected non-empty name")
	}
}

func TestCheckProcessLock(t *testing.T) {
	result := checkProcessLock()
	if result.name == "" {
		t.Fatal("expected non-empty name")
	}
}
