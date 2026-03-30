package util

import (
	"testing"
	"time"

	"github.com/708u/slack-cli/internal/tz"
)

func TestFormatUnixTimestamp(t *testing.T) {
	tz.Set(time.UTC)
	defer tz.Set(time.Local)
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

func TestFormatUnixTimestamp_NonUTC(t *testing.T) {
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("failed to load Asia/Tokyo: %v", err)
	}
	tz.Set(tokyo)
	defer tz.Set(time.Local)

	// 1600000000 = 2020-09-13 12:26:40 UTC = 2020-09-13 21:26:40 JST
	got := FormatUnixTimestamp(1600000000)
	if got != "2020-09-13" {
		t.Errorf("FormatUnixTimestamp(1600000000) = %q, want %q", got, "2020-09-13")
	}

	// 1600084800 = 2020-09-14 12:00:00 UTC = 2020-09-14 21:00:00 JST
	got = FormatUnixTimestamp(1600084800)
	if got != "2020-09-14" {
		t.Errorf("FormatUnixTimestamp(1600084800) = %q, want %q", got, "2020-09-14")
	}
}

func TestFormatTimestampFixed(t *testing.T) {
	tz.Set(time.UTC)
	defer tz.Set(time.Local)
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

func TestFormatTimestampFixed_NonUTC(t *testing.T) {
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("failed to load Asia/Tokyo: %v", err)
	}
	tz.Set(tokyo)
	defer tz.Set(time.Local)

	got := FormatTimestampFixed("1600000000.000000")
	want := "2020-09-13 21:26:40"
	if got != want {
		t.Errorf("FormatTimestampFixed = %q, want %q", got, want)
	}
}

func TestFormatSlackTimestamp(t *testing.T) {
	tz.Set(time.UTC)
	defer tz.Set(time.Local)

	got := FormatSlackTimestamp("1600000000.000000")
	want := "9/13/2020, 12:26:40 PM"
	if got != want {
		t.Errorf("FormatSlackTimestamp = %q, want %q", got, want)
	}

	got = FormatSlackTimestamp("invalid")
	if got != "invalid" {
		t.Errorf("expected fallback 'invalid', got %q", got)
	}
}
