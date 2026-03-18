package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// DeleteCmd deletes a sent message.
type DeleteCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	TS      string `name:"ts" required:"" help:"Message timestamp to delete"`
	Profile string `name:"profile" help:"Use specific workspace profile" optional:""`
}

// Run executes the delete command.
func (c *DeleteCmd) Run() error {
	if !threadTSPattern.MatchString(c.TS) {
		return fmt.Errorf("invalid message timestamp format")
	}

	token, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(token)
	if err := client.DeleteMessage(c.Channel, c.TS); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Message deleted successfully from #%s", c.Channel))
	return nil
}
