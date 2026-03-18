package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// LeaveCmd removes the authenticated user from a channel.
type LeaveCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the leave command.
func (c *LeaveCmd) Run() error {
	token, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(token)
	if err := client.LeaveChannel(c.Channel); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Left channel #%s", c.Channel))
	return nil
}
