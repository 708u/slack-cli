package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// EditCmd edits a sent message.
type EditCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	TS      string `name:"ts" required:"" help:"Message timestamp to edit"`
	Message string `name:"message" short:"m" required:"" help:"New message text"`
}

// Run executes the edit command.
func (c *EditCmd) Run(client *slack.Client) error {
	if !threadTSPattern.MatchString(c.TS) {
		return fmt.Errorf("invalid message timestamp format")
	}

	if err := client.UpdateMessage(c.Channel, c.TS, c.Message); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Message updated successfully in #%s", c.Channel))
	return nil
}
