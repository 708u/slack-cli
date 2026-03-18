// Package util provides common utility functions for the slack-cli,
// including terminal text sanitization, date formatting, Slack mention
// parsing, channel helpers, and shared error types.
package util

import (
	"regexp"
	"strings"
)

// oscSequencePattern matches OSC escape sequences (ESC ] ... BEL/ST).
var oscSequencePattern = regexp.MustCompile("\x1b\\][^\x07\x1b]*(?:\x07|\x1b\\\\)")

// ansiSequencePattern matches ANSI CSI and single-character escape sequences.
var ansiSequencePattern = regexp.MustCompile("\x1b(?:[@-Z\\\\\\-_]|\\[[0-?]*[ -/]*[@-~])")

// SanitizeTerminalText removes control characters from s to prevent
// terminal escape injection. Tab (0x09) and newline (0x0A) are kept for
// readability.
func SanitizeTerminalText(s string) string {
	if s == "" {
		return ""
	}

	// Strip ANSI / OSC escape sequences first.
	cleaned := oscSequencePattern.ReplaceAllString(s, "")
	cleaned = ansiSequencePattern.ReplaceAllString(cleaned, "")

	var b strings.Builder
	b.Grow(len(cleaned))

	for _, r := range cleaned {
		switch {
		case r == '\t', r == '\n':
			b.WriteRune(r)
		case r < 0x20, r == 0x7f, r >= 0x80 && r <= 0x9f:
			// Drop ASCII C0 controls, DEL, and C1 controls.
			continue
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// SanitizeTerminalData recursively sanitizes all string values found in
// v. Supported container types are map[string]any and []any; all other
// types are returned unchanged.
func SanitizeTerminalData(v any) any {
	switch val := v.(type) {
	case string:
		return SanitizeTerminalText(val)
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = SanitizeTerminalData(item)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, nested := range val {
			out[k] = SanitizeTerminalData(nested)
		}
		return out
	default:
		return v
	}
}
