package util

import "regexp"

// UserMentionPattern matches Slack user mentions in the format
// <@USERID>.
var UserMentionPattern = regexp.MustCompile(`<@([A-Z0-9]+)>`)

// MessageLike is the minimal interface a message must satisfy for
// ExtractAllUserIDs to read its author and body.
type MessageLike interface {
	GetUser() string
	GetText() string
}

// ExtractUserIDsFromMentions returns all user IDs referenced via
// <@USERID> mentions in text. The returned slice may contain
// duplicates.
func ExtractUserIDsFromMentions(text string) []string {
	matches := UserMentionPattern.FindAllStringSubmatch(text, -1)
	ids := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			ids = append(ids, m[1])
		}
	}
	return ids
}

// ExtractAllUserIDs collects every unique user ID found across
// messages, including both message authors and mentioned users.
func ExtractAllUserIDs(messages []MessageLike) []string {
	seen := make(map[string]struct{})
	var ids []string

	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	for _, msg := range messages {
		add(msg.GetUser())
		for _, id := range ExtractUserIDsFromMentions(msg.GetText()) {
			add(id)
		}
	}
	return ids
}

// FormatMessageWithMentions replaces <@USERID> mentions in message with
// @username using the provided user-ID-to-name map. Both the message
// body and resolved usernames are sanitized for terminal output.
func FormatMessageWithMentions(message string, users map[string]string) string {
	sanitized := SanitizeTerminalText(message)
	return UserMentionPattern.ReplaceAllStringFunc(sanitized, func(match string) string {
		sub := UserMentionPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		userID := sub[1]
		name, ok := users[userID]
		if !ok {
			name = userID
		}
		return "@" + SanitizeTerminalText(name)
	})
}

// ResolveUsername returns a display name for a message author. It looks
// up user in the users map first; if absent and botID is non-empty it
// returns "Bot"; otherwise "Unknown".
func ResolveUsername(user, botID string, users map[string]string) string {
	if user != "" {
		if name, ok := users[user]; ok {
			return SanitizeTerminalText(name)
		}
		return "Unknown User"
	}
	if botID != "" {
		return "Bot"
	}
	return "Unknown"
}
