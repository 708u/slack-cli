package cmd

import (
	"fmt"

	"github.com/fatih/color"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/708u/slack-cli/internal/util"
)

// ScheduledCmd manages scheduled messages (list, cancel).
type ScheduledCmd struct {
	List   ScheduledListCmd   `cmd:"" help:"List scheduled messages"`
	Cancel ScheduledCancelCmd `cmd:"" help:"Cancel a scheduled message"`
}

// ScheduledListCmd lists scheduled messages.
type ScheduledListCmd struct {
	Channel string `name:"channel" short:"c" help:"Filter by channel name or ID" optional:""`
	Limit   int    `name:"limit" help:"Maximum number of scheduled messages to list" default:"50"`
	Format  string `name:"format" enum:"table,simple,json" default:"table"`
	Profile string `name:"profile" help:"Workspace profile" optional:""`
}

// Run executes the scheduled list command.
func (c *ScheduledListCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)

	messages, err := client.ListScheduledMessages(c.Channel, c.Limit)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		fmt.Println("No scheduled messages found")
		return nil
	}

	infos := make([]format.ScheduledMessageInfo, len(messages))
	for i, m := range messages {
		infos[i] = format.ScheduledMessageInfo{
			ID:        util.SanitizeTerminalText(m.ID),
			ChannelID: util.SanitizeTerminalText(m.ChannelID),
			PostAt:    format.FormatPostAt(m.PostAt),
			Text:      util.SanitizeTerminalText(m.Text),
		}
	}

	f := format.ParseFormat(c.Format)
	format.FormatScheduledMessages(infos, f)
	return nil
}

// ScheduledCancelCmd cancels a scheduled message.
type ScheduledCancelCmd struct {
	Channel string `name:"channel" short:"c" help:"Channel name or ID" required:""`
	ID      string `name:"id" help:"Scheduled message ID" required:""`
	Profile string `name:"profile" help:"Workspace profile" optional:""`
}

// Run executes the scheduled cancel command.
func (c *ScheduledCancelCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)

	if err := client.CancelScheduledMessage(c.Channel, c.ID); err != nil {
		return err
	}

	green := color.New(color.FgGreen)
	green.Printf("Scheduled message %s cancelled\n", c.ID)
	return nil
}
