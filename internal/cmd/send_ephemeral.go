package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// SendEphemeralCmd sends an ephemeral message visible only to a
// specific user in a channel.
type SendEphemeralCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Target channel name or ID"`
	User    string `name:"user" short:"u" required:"" help:"User ID who will see the ephemeral message"`
	Message string `name:"message" short:"m" required:"" help:"Message to send"`
	Thread  string `name:"thread" short:"t" help:"Thread timestamp to reply to" optional:""`
	Profile string `name:"profile" help:"Use specific workspace profile" optional:""`
}

// Run executes the send-ephemeral command.
func (c *SendEphemeralCmd) Run() error {
	if c.Thread != "" && !threadTSPattern.MatchString(c.Thread) {
		return fmt.Errorf("invalid thread timestamp format")
	}

	token, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(token)
	if err := client.SendEphemeralMessage(c.Channel, c.User, c.Message, c.Thread); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Ephemeral message sent to #%s", c.Channel))
	return nil
}
