package format

import (
	"fmt"
	"time"

	"github.com/708u/slack-cli/internal/tz"
	"github.com/fatih/color"
)

// PinInfo holds the fields needed to display a pinned item.
type PinInfo struct {
	Type      string
	Channel   string
	Text      string
	User      string
	MessageTS string
	Created   int64
}

// FormatPins prints a list of pinned items in the requested format.
func FormatPins(items []PinInfo, f Format) {
	switch f {
	case JSON:
		type jsonPin struct {
			Type      string `json:"type"`
			Channel   string `json:"channel"`
			Text      string `json:"text"`
			User      string `json:"user"`
			MessageTS string `json:"message_ts"`
			Created   int64  `json:"created"`
		}
		out := make([]jsonPin, len(items))
		for i, p := range items {
			out[i] = jsonPin{
				Type:      p.Type,
				Channel:   p.Channel,
				Text:      p.Text,
				User:      p.User,
				MessageTS: p.MessageTS,
				Created:   p.Created,
			}
		}
		PrintJSON(out)
	case Simple:
		for _, p := range items {
			fmt.Printf("%s\t%s\t%s\n", p.Channel, p.User, textOrDefault(p.Text))
		}
	default:
		bold := color.New(color.Bold)
		bold.Printf("%-16s%-14s%-12s%s\n", "Channel", "User", "Created", "Text")
		fmt.Println(Separator(70))
		for _, p := range items {
			created := formatUnixDate(p.Created)
			text := truncate(textOrDefault(p.Text), 35)
			fmt.Printf("%-16s%-14s%-12s%s\n", p.Channel, p.User, created, text)
		}
	}
}

// ReminderInfo holds the fields needed to display a reminder.
type ReminderInfo struct {
	ID         string
	Text       string
	Time       int64
	CompleteTS int64
	Recurring  bool
}

// FormatReminders prints a list of reminders in the requested format.
func FormatReminders(reminders []ReminderInfo, f Format) {
	switch f {
	case JSON:
		type jsonReminder struct {
			ID            string `json:"id"`
			Text          string `json:"text"`
			Time          int64  `json:"time"`
			TimeFormatted string `json:"time_formatted"`
			Status        string `json:"status"`
			Recurring     bool   `json:"recurring"`
		}
		out := make([]jsonReminder, len(reminders))
		for i, r := range reminders {
			out[i] = jsonReminder{
				ID:            r.ID,
				Text:          r.Text,
				Time:          r.Time,
				TimeFormatted: FormatUnixISO(r.Time),
				Status:        reminderStatus(r.CompleteTS),
				Recurring:     r.Recurring,
			}
		}
		PrintJSON(out)
	case Simple:
		for _, r := range reminders {
			fmt.Printf("%s\t%s\t%s\t%s\n",
				r.ID, r.Text, FormatUnixISO(r.Time), reminderStatus(r.CompleteTS))
		}
	default:
		const (
			idW     = 14
			textW   = 30
			timeW   = 26
			statusW = 10
		)
		bold := color.New(color.Bold)
		bold.Printf("%-*s%-*s%-*s%-*s\n", idW, "ID", textW, "Text", timeW, "Time", statusW, "Status")
		fmt.Println(Separator(idW + textW + timeW + statusW))
		for _, r := range reminders {
			text := truncate(r.Text, textW-2)
			fmt.Printf("%-*s%-*s%-*s%-*s\n",
				idW, r.ID,
				textW, text,
				timeW, FormatUnixISO(r.Time),
				statusW, reminderStatus(r.CompleteTS))
		}
	}
}

// BookmarkInfo holds the fields needed to display a bookmarked (starred) item.
type BookmarkInfo struct {
	Type       string
	Channel    string
	Text       string
	MessageTS  string
	DateCreate int64
}

// FormatBookmarks prints a list of bookmarked items in the requested format.
func FormatBookmarks(items []BookmarkInfo, f Format) {
	switch f {
	case JSON:
		type jsonBookmark struct {
			Type       string `json:"type"`
			Channel    string `json:"channel"`
			MessageTS  string `json:"timestamp"`
			Text       string `json:"text"`
			DateCreate int64  `json:"date_create"`
			SavedAt    string `json:"saved_at"`
		}
		out := make([]jsonBookmark, len(items))
		for i, b := range items {
			out[i] = jsonBookmark{
				Type:       b.Type,
				Channel:    b.Channel,
				MessageTS:  b.MessageTS,
				Text:       b.Text,
				DateCreate: b.DateCreate,
				SavedAt:    FormatUnixISO(b.DateCreate),
			}
		}
		PrintJSON(out)
	case Simple:
		for _, b := range items {
			fmt.Printf("%s\t%s\t%s\t%s\n",
				b.Channel, b.MessageTS, b.Text, FormatUnixISO(b.DateCreate))
		}
	default:
		const (
			chW   = 16
			tsW   = 20
			textW = 40
			dateW = 26
		)
		bold := color.New(color.Bold)
		bold.Printf("%-*s%-*s%-*s%-*s\n", chW, "Channel", tsW, "Timestamp", textW, "Text", dateW, "Saved At")
		fmt.Println(Separator(chW + tsW + textW + dateW))
		for _, b := range items {
			text := truncate(b.Text, textW-2)
			fmt.Printf("%-*s%-*s%-*s%-*s\n",
				chW, b.Channel,
				tsW, b.MessageTS,
				textW, text,
				dateW, FormatUnixISO(b.DateCreate))
		}
	}
}

func reminderStatus(completeTS int64) string {
	if completeTS > 0 {
		return "completed"
	}
	return "pending"
}

// FormatUnixISO formats a Unix epoch as an RFC3339 string in the
// configured timezone. Returns "" for zero.
func FormatUnixISO(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).In(tz.Location()).Format(time.RFC3339)
}

func formatUnixDate(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).In(tz.Location()).Format("2006-01-02")
}
