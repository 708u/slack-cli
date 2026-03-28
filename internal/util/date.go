package util

import (
	"fmt"
	"strconv"
	"time"

	"github.com/708u/slack-cli/internal/tz"
)

// FormatUnixTimestamp formats a Unix epoch (seconds) as "YYYY-MM-DD" in
// the configured timezone.
func FormatUnixTimestamp(ts int64) string {
	t := time.Unix(ts, 0).In(tz.Location())
	return t.Format("2006-01-02")
}

// FormatSlackTimestamp parses a Slack-style timestamp string (e.g.
// "1234567890.123456") and returns a locale-like string representation
// in the local timezone.
func FormatSlackTimestamp(ts string) string {
	f, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return ts
	}
	sec := int64(f)
	nsec := int64((f - float64(sec)) * 1e9)
	t := time.Unix(sec, nsec).In(tz.Location())
	return t.Format("1/2/2006, 3:04:05 PM")
}

// FormatTimestampFixed parses a Slack-style timestamp string and returns
// "YYYY-MM-DD HH:MM:SS" in the configured timezone.
func FormatTimestampFixed(ts string) string {
	f, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return ts
	}
	sec := int64(f)
	nsec := int64((f - float64(sec)) * 1e9)
	t := time.Unix(sec, nsec).In(tz.Location())
	return fmt.Sprintf(
		"%04d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
	)
}
