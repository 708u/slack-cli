package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/708u/slack-cli/internal/util"
)

// HistoryCmd retrieves message history from a Slack channel or DM.
type HistoryCmd struct {
	Channel  string `name:"channel" short:"c" help:"Channel name or ID" optional:""`
	User     string `name:"user" short:"u" help:"Show DM history by username" optional:""`
	UserID   string `name:"user-id" help:"Show DM history by user ID" optional:""`
	Number   int    `name:"number" short:"n" help:"Number of messages" default:"10"`
	Since    string `name:"since" help:"Messages since date (YYYY-MM-DD HH:MM:SS)" optional:""`
	Thread   string `name:"thread" short:"t" help:"Thread timestamp" optional:""`
	WithLink bool   `name:"with-link" help:"Include permalink URLs"`
	Format   string `name:"format" enum:"table,simple,json" default:"table"`
}

// Run executes the history command.
func (c *HistoryCmd) Run(client *slack.Client) error {
	targets := 0
	if c.Channel != "" {
		targets++
	}
	if c.User != "" {
		targets++
	}
	if c.UserID != "" {
		targets++
	}
	if targets == 0 {
		return fmt.Errorf("you must specify one of: --channel, --user, or --user-id")
	}
	if targets > 1 {
		return fmt.Errorf("--channel, --user, and --user-id are mutually exclusive")
	}

	channelID, displayName, err := c.resolveTarget(client)
	if err != nil {
		return err
	}

	var result *slack.HistoryResult

	if c.Thread != "" {
		if c.Number != 10 {
			fmt.Println("Warning: --number is ignored when --thread is specified.")
		}
		if c.Since != "" {
			fmt.Println("Warning: --since is ignored when --thread is specified.")
		}
		result, err = client.GetThreadHistory(channelID, c.Thread)
	} else {
		opts := slack.HistoryOptions{
			Limit: c.Number,
		}

		oldest, parseErr := prepareSinceTimestamp(c.Since)
		if parseErr != nil {
			return parseErr
		}
		if oldest != "" {
			opts.Oldest = oldest
		}

		result, err = client.GetHistory(channelID, opts)
	}
	if err != nil {
		return err
	}

	// Reverse messages for chronological order unless viewing a thread.
	messages := result.Messages
	if c.Thread == "" {
		messages = reverseMessages(messages)
	}

	// Optionally fetch permalinks.
	var permalinks map[string]string
	if c.WithLink && len(messages) > 0 {
		timestamps := make([]string, len(messages))
		for i, m := range messages {
			timestamps[i] = m.TS
		}
		permalinks, _ = client.GetPermalinks(channelID, timestamps)
	}

	infos := make([]format.MessageInfo, len(messages))
	for i, m := range messages {
		username := util.ResolveUsername(m.User, m.BotID, result.Users)
		text := util.FormatMessageWithMentions(m.Text, result.Users)
		var permalink string
		if permalinks != nil {
			permalink = permalinks[m.TS]
		}
		infos[i] = format.MessageInfo{
			TS:         m.TS,
			Timestamp:  util.FormatSlackTimestamp(m.TS),
			Username:   username,
			Text:       text,
			Permalink:  permalink,
			ThreadTS:   m.ThreadTS,
			ReplyCount: m.ReplyCount,
		}
	}

	f := format.ParseFormat(c.Format)
	format.FormatHistory(format.HistoryFormatOpts{
		ChannelName: displayName,
		Messages:    infos,
	}, f)

	return nil
}

func (c *HistoryCmd) resolveTarget(client *slack.Client) (channelID, displayName string, err error) {
	if c.User != "" {
		userID, err := client.ResolveUserIDByName(c.User)
		if err != nil {
			return "", "", err
		}
		ch, err := client.OpenDMChannel(userID)
		if err != nil {
			return "", "", err
		}
		return ch, "@" + strings.TrimPrefix(c.User, "@"), nil
	}
	if c.UserID != "" {
		ch, err := client.OpenDMChannel(c.UserID)
		if err != nil {
			return "", "", err
		}
		return ch, c.UserID, nil
	}
	ch, err := client.ResolveChannelID(c.Channel)
	if err != nil {
		return "", "", err
	}
	return ch, c.Channel, nil
}

// prepareSinceTimestamp converts a date string (YYYY-MM-DD HH:MM:SS) to
// a Unix timestamp string. Returns ("", nil) when since is empty.
func prepareSinceTimestamp(since string) (string, error) {
	if since == "" {
		return "", nil
	}

	t, err := time.Parse("2006-01-02 15:04:05", since)
	if err != nil {
		return "", fmt.Errorf("invalid date format %q: use YYYY-MM-DD HH:MM:SS", since)
	}

	ts := t.Unix()
	return strconv.FormatInt(ts, 10), nil
}

// reverseMessages returns a new slice with messages in reverse order.
func reverseMessages(msgs []slack.Message) []slack.Message {
	n := len(msgs)
	out := make([]slack.Message, n)
	for i, m := range msgs {
		out[n-1-i] = m
	}
	return out
}
