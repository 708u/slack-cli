package util

import "regexp"

// userIDPattern matches Slack user IDs, which start with U (standard
// workspaces) or W (Enterprise Grid) followed by 8 or more uppercase
// alphanumeric characters.
var userIDPattern = regexp.MustCompile(`^[UW][A-Z0-9]{8,}$`)

// IsUserID reports whether s looks like a Slack user ID.
func IsUserID(s string) bool {
	return userIDPattern.MatchString(s)
}
