package util

import "testing"

func TestSanitizeTerminalText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"plain text", "hello world", "hello world"},
		{"preserves tabs", "col1\tcol2", "col1\tcol2"},
		{"preserves newlines", "line1\nline2", "line1\nline2"},
		{"strips NUL", "hel\x00lo", "hello"},
		{"strips BEL", "he\x07llo", "hello"},
		{"strips ESC", "he\x1bllo", "hello"},
		{"strips DEL", "hel\x7flo", "hello"},
		{"strips C1 controls", "hel\xc2\x80\xc2\x9flo", "hello"},
		{"strips ANSI color", "\x1b[31mred\x1b[0m", "red"},
		{"strips OSC sequence", "\x1b]0;title\x07text", "text"},
		{"unicode preserved", "Hello", "Hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeTerminalText(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeTerminalText(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeTerminalData(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		got := SanitizeTerminalData("he\x00llo")
		if got != "hello" {
			t.Errorf("expected 'hello', got %q", got)
		}
	})

	t.Run("slice", func(t *testing.T) {
		input := []any{"he\x00llo", "wo\x07rld"}
		got := SanitizeTerminalData(input).([]any)
		if got[0] != "hello" || got[1] != "world" {
			t.Errorf("unexpected: %v", got)
		}
	})

	t.Run("map", func(t *testing.T) {
		input := map[string]any{"key": "va\x00lue"}
		got := SanitizeTerminalData(input).(map[string]any)
		if got["key"] != "value" {
			t.Errorf("unexpected: %v", got)
		}
	})

	t.Run("int passthrough", func(t *testing.T) {
		got := SanitizeTerminalData(42)
		if got != 42 {
			t.Errorf("expected 42, got %v", got)
		}
	})
}
