package cmd

import (
	"fmt"

	"github.com/fatih/color"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/708u/slack-cli/internal/util"
)

// UnreadCmd shows unread messages across channels.
type UnreadCmd struct {
	Channel   string `name:"channel" short:"c" help:"Specific channel" optional:""`
	Format    string `name:"format" enum:"table,simple,json" default:"table"`
	CountOnly bool   `name:"count-only" help:"Show only counts"`
	Limit     int    `name:"limit" help:"Max channels" default:"50"`
	MarkRead  bool   `name:"mark-read" help:"Mark as read after fetching"`
	Profile   string `name:"profile" optional:""`
}

// Run executes the unread command.
func (c *UnreadCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)

	if c.Channel != "" {
		return c.handleSpecificChannel(client)
	}
	return c.handleAllChannels(client)
}

// handleSpecificChannel shows unread messages for a single channel.
func (c *UnreadCmd) handleSpecificChannel(client *slack.Client) error {
	result, err := client.GetChannelUnread(c.Channel)
	if err != nil {
		return err
	}

	infos := make([]format.MessageInfo, len(result.Messages))
	for i, m := range result.Messages {
		username := util.ResolveUsername(m.User, m.BotID, result.Users)
		text := util.FormatMessageWithMentions(m.Text, result.Users)
		infos[i] = format.MessageInfo{
			TS:        m.TS,
			Timestamp: util.FormatSlackTimestamp(m.TS),
			Username:  username,
			Text:      text,
		}
	}

	f := format.ParseFormat(c.Format)
	format.FormatUnreadMessages(format.UnreadFormatOpts{
		ChannelName:           util.SanitizeTerminalText(result.Channel.Name),
		ChannelID:             result.Channel.ID,
		UnreadCount:           result.TotalUnreadCount,
		Messages:              infos,
		CountOnly:             c.CountOnly,
		DisplayedMessageCount: result.DisplayedMessageCount,
	}, f)

	if c.MarkRead {
		if err := client.MarkAsRead(result.Channel.ID); err != nil {
			return err
		}
		green := color.New(color.FgGreen)
		green.Printf("Marked messages in %s as read\n",
			util.FormatChannelName(result.Channel.Name))
	}

	return nil
}

// handleAllChannels shows channels with unread messages.
func (c *UnreadCmd) handleAllChannels(client *slack.Client) error {
	channels, err := client.ListUnreadChannels()
	if err != nil {
		return err
	}

	if len(channels) == 0 {
		green := color.New(color.FgGreen)
		green.Println("No unread messages")
		return nil
	}

	display := channels
	if c.Limit > 0 && len(display) > c.Limit {
		display = display[:c.Limit]
	}

	infos := make([]format.UnreadChannelInfo, len(display))
	for i, ch := range display {
		infos[i] = format.UnreadChannelInfo{
			ID:          ch.ID,
			Name:        util.FormatChannelName(ch.Name),
			UnreadCount: ch.UnreadCountDisp,
			LastRead:    ch.LastRead,
		}
	}

	f := format.ParseFormat(c.Format)
	format.FormatUnreadChannels(infos, f, c.CountOnly)

	if c.MarkRead {
		for _, ch := range channels {
			if err := client.MarkAsRead(ch.ID); err != nil {
				return fmt.Errorf("failed to mark %s as read: %w", ch.ID, err)
			}
		}
		green := color.New(color.FgGreen)
		green.Println("Marked all messages as read")
	}

	return nil
}
