package format

import (
	"fmt"

	"github.com/fatih/color"
)

// MessageInfo holds the fields needed to display a single message.
type MessageInfo struct {
	TS         string
	Timestamp  string
	Username   string
	Text       string
	Permalink  string
	ThreadTS   string
	ReplyCount int
}

// HistoryFormatOpts holds options for formatting channel history output.
type HistoryFormatOpts struct {
	ChannelName string
	Messages    []MessageInfo
}

// FormatHistory prints channel message history in the requested format.
func FormatHistory(opts HistoryFormatOpts, f Format) {
	switch f {
	case JSON:
		type jsonMsg struct {
			TS         string `json:"ts"`
			Timestamp  string `json:"timestamp"`
			User       string `json:"user"`
			Text       string `json:"text"`
			ThreadTS   string `json:"thread_ts,omitempty"`
			ReplyCount int    `json:"reply_count,omitempty"`
			Permalink  string `json:"permalink,omitempty"`
		}
		type jsonOut struct {
			Channel  string    `json:"channel"`
			Messages []jsonMsg `json:"messages"`
			Total    int       `json:"total"`
		}
		msgs := make([]jsonMsg, len(opts.Messages))
		for i, m := range opts.Messages {
			msgs[i] = jsonMsg{
				TS:         m.TS,
				Timestamp:  m.Timestamp,
				User:       m.Username,
				Text:       textOrDefault(m.Text),
				ThreadTS:   m.ThreadTS,
				ReplyCount: m.ReplyCount,
				Permalink:  m.Permalink,
			}
		}
		PrintJSON(jsonOut{
			Channel:  opts.ChannelName,
			Messages: msgs,
			Total:    len(opts.Messages),
		})
	case Simple:
		if len(opts.Messages) == 0 {
			fmt.Println("No messages found")
			return
		}
		for _, m := range opts.Messages {
			line := fmt.Sprintf("[%s] %s: %s", m.Timestamp, m.Username, textOrDefault(m.Text))
			if m.Permalink != "" {
				line += " " + m.Permalink
			}
			fmt.Println(line)
		}
	default:
		bold := color.New(color.Bold)
		gray := color.New(color.FgHiBlack)
		cyan := color.New(color.FgCyan)
		blue := color.New(color.FgBlue)
		green := color.New(color.FgGreen)

		bold.Printf("\nMessage History for #%s:\n", opts.ChannelName)

		if len(opts.Messages) == 0 {
			color.Yellow("No messages found")
			return
		}

		fmt.Println()
		for _, m := range opts.Messages {
			fmt.Printf("%s %s\n", gray.Sprintf("[%s]", m.Timestamp), cyan.Sprint(m.Username))
			fmt.Println(textOrDefault(m.Text))
			if m.Permalink != "" {
				blue.Println(m.Permalink)
			}
			fmt.Println()
		}

		green.Printf("Displayed %d message(s)\n", len(opts.Messages))
	}
}

// UnreadFormatOpts holds options for formatting unread messages.
type UnreadFormatOpts struct {
	ChannelName           string
	ChannelID             string
	UnreadCount           int
	Messages              []MessageInfo
	CountOnly             bool
	DisplayedMessageCount int
}

// FormatUnreadMessages prints unread messages for a channel.
func FormatUnreadMessages(opts UnreadFormatOpts, f Format) {
	switch f {
	case JSON:
		type jsonMsg struct {
			Timestamp string `json:"timestamp"`
			Author    string `json:"author"`
			Text      string `json:"text"`
		}
		type jsonOut struct {
			Channel               string    `json:"channel"`
			ChannelID             string    `json:"channel_id"`
			UnreadCount           int       `json:"unread_count"`
			DisplayedMessageCount int       `json:"displayed_message_count,omitempty"`
			IsTruncated           bool      `json:"is_truncated,omitempty"`
			Messages              []jsonMsg `json:"messages,omitempty"`
		}
		out := jsonOut{
			Channel:     opts.ChannelName,
			ChannelID:   opts.ChannelID,
			UnreadCount: opts.UnreadCount,
		}
		if !opts.CountOnly && len(opts.Messages) > 0 {
			msgs := make([]jsonMsg, len(opts.Messages))
			for i, m := range opts.Messages {
				msgs[i] = jsonMsg{
					Timestamp: m.Timestamp,
					Author:    m.Username,
					Text:      textOrDefault(m.Text),
				}
			}
			out.Messages = msgs
			if opts.DisplayedMessageCount > 0 && opts.DisplayedMessageCount < opts.UnreadCount {
				out.DisplayedMessageCount = opts.DisplayedMessageCount
				out.IsTruncated = true
			}
		}
		PrintJSON(out)
	case Simple:
		fmt.Printf("%s (%d)\n", opts.ChannelName, opts.UnreadCount)
		if !opts.CountOnly && len(opts.Messages) > 0 {
			for _, m := range opts.Messages {
				fmt.Printf("[%s] %s: %s\n", m.Timestamp, m.Username, textOrDefault(m.Text))
			}
			if opts.DisplayedMessageCount > 0 && opts.DisplayedMessageCount < opts.UnreadCount {
				fmt.Printf("Showing latest %d of %d unread messages\n",
					opts.DisplayedMessageCount, opts.UnreadCount)
			}
		}
	default:
		bold := color.New(color.Bold)
		gray := color.New(color.FgHiBlack)
		cyan := color.New(color.FgCyan)

		bold.Printf("%s: %d unread messages\n", opts.ChannelName, opts.UnreadCount)

		if !opts.CountOnly && len(opts.Messages) > 0 {
			fmt.Println()
			for _, m := range opts.Messages {
				fmt.Printf("%s %s\n", gray.Sprint(m.Timestamp), cyan.Sprint(m.Username))
				fmt.Println(textOrDefault(m.Text))
				fmt.Println()
			}
			if opts.DisplayedMessageCount > 0 && opts.DisplayedMessageCount < opts.UnreadCount {
				gray.Printf("Showing latest %d of %d unread messages\n",
					opts.DisplayedMessageCount, opts.UnreadCount)
			}
		}
	}
}

func textOrDefault(s string) string {
	if s == "" {
		return "(no text)"
	}
	return s
}
