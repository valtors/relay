// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Tamish Max
package tools

import (
	"testing"
)

func TestFormatJSONQueryValue_Nil(t *testing.T) {
	if got := formatJSONQueryValue(nil); got != "null" {
		t.Errorf("expected null, got %s", got)
	}
}

func TestFormatJSONQueryValue_String(t *testing.T) {
	if got := formatJSONQueryValue("hello"); got != "hello" {
		t.Errorf("expected hello, got %s", got)
	}
}

func TestFormatJSONQueryValue_Bool(t *testing.T) {
	if got := formatJSONQueryValue(true); got != "true" {
		t.Errorf("expected true, got %s", got)
	}
	if got := formatJSONQueryValue(false); got != "false" {
		t.Errorf("expected false, got %s", got)
	}
}

func TestFormatJSONQueryValue_Float(t *testing.T) {
	if got := formatJSONQueryValue(3.14); got != "3.14" {
		t.Errorf("expected 3.14, got %s", got)
	}
}

func TestFormatJSONQueryValue_Int(t *testing.T) {
	if got := formatJSONQueryValue(42); got != "42" {
		t.Errorf("expected 42, got %s", got)
	}
}

func TestFormatJSONQueryValue_Slice(t *testing.T) {
	got := formatJSONQueryValue([]int{1, 2, 3})
	if got != "[1,2,3]" {
		t.Errorf("expected [1,2,3], got %s", got)
	}
}

func TestFormatJSONQueryValue_Map(t *testing.T) {
	got := formatJSONQueryValue(map[string]int{"a": 1})
	if got != `{"a":1}` {
		t.Errorf("expected {\"a\":1}, got %s", got)
	}
}

func TestCSVValue(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{nil, ""},
		{"hello", "hello"},
		{true, "true"},
		{3.14, "3.14"},
	}
	for _, tt := range tests {
		got := csvValue(tt.input)
		if got != tt.expected {
			t.Errorf("csvValue(%v) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestMaxInt(t *testing.T) {
	if maxInt(3, 7) != 7 {
		t.Error("expected 7")
	}
	if maxInt(10, 2) != 10 {
		t.Error("expected 10")
	}
	if maxInt(-1, 0) != 0 {
		t.Error("expected 0")
	}
}

func TestValidateURL_Valid(t *testing.T) {
	tests := []string{
		"http://example.com",
		"https://example.com/path",
		"https://example.com/path?q=1",
	}
	for _, url := range tests {
		if err := validateURL(url); err != nil {
			t.Errorf("validateURL(%q) = %v, want nil", url, err)
		}
	}
}

func TestValidateURL_Invalid(t *testing.T) {
	tests := []string{
		"not a url",
		"ftp://example.com",
		"javascript:alert(1)",
		"",
	}
	for _, url := range tests {
		if err := validateURL(url); err == nil {
			t.Errorf("validateURL(%q) = nil, want error", url)
		}
	}
}
