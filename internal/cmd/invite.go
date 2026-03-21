package cmd

import (
	"fmt"
	"strings"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// InviteCmd invites user(s) to a channel.
type InviteCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	Users   string `name:"users" short:"u" required:"" help:"Comma-separated user IDs to invite"`
	Force   bool   `name:"force" optional:"" help:"Continue inviting valid users even if some IDs are invalid"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the invite command.
func (c *InviteCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	var userIDs []string
	for id := range strings.SplitSeq(c.Users, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			userIDs = append(userIDs, id)
		}
	}
	if len(userIDs) == 0 {
		return fmt.Errorf("at least one valid user ID is required")
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	if err := client.InviteToChannel(c.Channel, userIDs, c.Force); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Invited user(s) to channel #%s", c.Channel))
	return nil
}
