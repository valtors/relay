package main

import (
	"testing"
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
