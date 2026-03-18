package util

import "regexp"

// channelIDPattern matches Slack channel IDs, which start with C, D, or
// G followed by 8 or more uppercase alphanumeric characters.
var channelIDPattern = regexp.MustCompile(`^[CDG][A-Z0-9]{8,}$`)

// IsChannelID reports whether s looks like a Slack channel ID.
func IsChannelID(s string) bool {
	return channelIDPattern.MatchString(s)
}

// FormatChannelName returns name prefixed with '#' if it is not already
// present. An empty name returns "#unknown".
func FormatChannelName(name string) string {
	if name == "" {
		return "#unknown"
	}
	sanitized := SanitizeTerminalText(name)
	if len(sanitized) > 0 && sanitized[0] == '#' {
		return sanitized
	}
	return "#" + sanitized
}

// channelTypeMap translates friendly type names to Slack API type
// strings.
var channelTypeMap = map[string]string{
	"public":  "public_channel",
	"private": "private_channel",
	"im":      "im",
	"mpim":    "mpim",
	"all":     "public_channel,private_channel,mpim,im",
}

// GetChannelTypes maps a human-readable channel type name (public,
// private, im, mpim, all) to the corresponding Slack API types string.
// Unrecognised values default to "public_channel".
func GetChannelTypes(t string) string {
	if v, ok := channelTypeMap[t]; ok {
		return v
	}
	return "public_channel"
}
