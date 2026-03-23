package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// JoinCmd joins the authenticated user to a channel.
type JoinCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
}

// Run executes the join command.
func (c *JoinCmd) Run(client *slack.Client) error {
	if err := client.JoinChannel(c.Channel); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Joined channel #%s", c.Channel))
	return nil
}
