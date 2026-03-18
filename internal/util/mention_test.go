package util

import "testing"

func TestExtractUserIDsFromMentions(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{"no mentions", "hello world", nil},
		{"single mention", "hi <@U001>", []string{"U001"}},
		{"multiple", "<@U001> and <@U002>", []string{"U001", "U002"}},
		{"embedded", "Hey <@U001>, look at <@U002>'s work", []string{"U001", "U002"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractUserIDsFromMentions(tt.text)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d IDs, got %d: %v", len(tt.expected), len(got), got)
			}
			for i, id := range got {
				if id != tt.expected[i] {
					t.Errorf("ID[%d] = %q, want %q", i, id, tt.expected[i])
				}
			}
		})
	}
}

type fakeMessage struct {
	user string
	text string
}

func (m fakeMessage) GetUser() string { return m.user }
func (m fakeMessage) GetText() string { return m.text }

func TestExtractAllUserIDs(t *testing.T) {
	msgs := []MessageLike{
		fakeMessage{user: "U001", text: "hello <@U002>"},
		fakeMessage{user: "U002", text: "reply <@U001> and <@U003>"},
		fakeMessage{user: "U001", text: "no mentions"},
	}

	ids := ExtractAllUserIDs(msgs)

	expected := map[string]bool{"U001": true, "U002": true, "U003": true}
	if len(ids) != len(expected) {
		t.Fatalf("expected %d unique IDs, got %d: %v", len(expected), len(ids), ids)
	}
	for _, id := range ids {
		if !expected[id] {
			t.Errorf("unexpected ID: %s", id)
		}
	}
}

func TestFormatMessageWithMentions(t *testing.T) {
	users := map[string]string{
		"U001": "alice",
		"U002": "bob",
	}
	got := FormatMessageWithMentions("Hi <@U001>, see <@U002>", users)
	expected := "Hi @alice, see @bob"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestFormatMessageWithMentions_UnknownUser(t *testing.T) {
	users := map[string]string{}
	got := FormatMessageWithMentions("Hi <@U999>", users)
	expected := "Hi @U999"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestResolveUsername(t *testing.T) {
	users := map[string]string{"U001": "alice"}

	tests := []struct {
		user, botID, expected string
	}{
		{"U001", "", "alice"},
		{"U999", "", "Unknown User"},
		{"", "B001", "Bot"},
		{"", "", "Unknown"},
	}
	for _, tt := range tests {
		got := ResolveUsername(tt.user, tt.botID, users)
		if got != tt.expected {
			t.Errorf("ResolveUsername(%q, %q) = %q, want %q", tt.user, tt.botID, got, tt.expected)
		}
	}
}
