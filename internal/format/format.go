package format

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Format specifies the output format for CLI results.
type Format string

const (
	Table  Format = "table"
	Simple Format = "simple"
	JSON   Format = "json"
)

// ParseFormat converts a string into a Format, defaulting to Table for
// unrecognised values.
func ParseFormat(s string) Format {
	switch Format(strings.ToLower(s)) {
	case Table:
		return Table
	case Simple:
		return Simple
	case JSON:
		return JSON
	default:
		return Table
	}
}

// PrintJSON marshals v as indented JSON and writes it to stdout.
func PrintJSON(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal error: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

// truncate shortens s to maxLen characters, appending "..." when truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Separator returns a horizontal line of the given width using the box-drawing
// character U+2500.
func Separator(width int) string {
	return strings.Repeat("\u2500", width)
}

// boolYesNo returns "Yes" or "No" for a boolean value.
func boolYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
