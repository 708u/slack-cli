package format

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

// ScheduledMessageInfo holds the fields needed to display a scheduled message.
type ScheduledMessageInfo struct {
	ID        string
	ChannelID string
	PostAt    string
	Text      string
}

// FormatScheduledMessages prints scheduled messages in the requested format.
func FormatScheduledMessages(messages []ScheduledMessageInfo, f Format) {
	switch f {
	case JSON:
		type jsonScheduled struct {
			ID        string `json:"id"`
			ChannelID string `json:"channel_id"`
			PostAt    string `json:"post_at"`
			Text      string `json:"text"`
		}
		out := make([]jsonScheduled, len(messages))
		for i, m := range messages {
			out[i] = jsonScheduled{
				ID:        m.ID,
				ChannelID: m.ChannelID,
				PostAt:    m.PostAt,
				Text:      m.Text,
			}
		}
		PrintJSON(out)
	case Simple:
		for _, m := range messages {
			fmt.Printf("%s %s %s %s\n", m.PostAt, m.ChannelID, m.ID, m.Text)
		}
	default:
		bold := color.New(color.Bold)
		bold.Printf("%-22s%-14s%-20s%s\n", "Scheduled At", "Channel", "ID", "Text")
		fmt.Println(Separator(70))
		for _, m := range messages {
			text := truncate(m.Text, 30)
			fmt.Printf("%-22s%-14s%-20s%s\n", m.PostAt, m.ChannelID, m.ID, text)
		}
	}
}

// FormatPostAt converts a Unix epoch (seconds) to an ISO 8601 string.
func FormatPostAt(postAt int64) string {
	return time.Unix(postAt, 0).UTC().Format(time.RFC3339)
}
