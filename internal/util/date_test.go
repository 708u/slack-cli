package util

import "testing"

func TestFormatUnixTimestamp(t *testing.T) {
	tests := []struct {
		ts       int64
		expected string
	}{
		{0, "1970-01-01"},
		{1600000000, "2020-09-13"},
		{1700000000, "2023-11-14"},
	}
	for _, tt := range tests {
		got := FormatUnixTimestamp(tt.ts)
		if got != tt.expected {
			t.Errorf("FormatUnixTimestamp(%d) = %q, want %q", tt.ts, got, tt.expected)
		}
	}
}

func TestFormatTimestampFixed(t *testing.T) {
	tests := []struct {
		ts       string
		expected string
	}{
		{"1600000000.000000", "2020-09-13 12:26:40"},
		{"invalid", "invalid"},
	}
	for _, tt := range tests {
		got := FormatTimestampFixed(tt.ts)
		if got != tt.expected {
			t.Errorf("FormatTimestampFixed(%q) = %q, want %q", tt.ts, got, tt.expected)
		}
	}
}

func TestFormatSlackTimestamp(t *testing.T) {
	// Just check it doesn't panic on valid/invalid input
	got := FormatSlackTimestamp("1600000000.000000")
	if got == "" {
		t.Error("expected non-empty result")
	}

	got = FormatSlackTimestamp("invalid")
	if got != "invalid" {
		t.Errorf("expected fallback 'invalid', got %q", got)
	}
}
