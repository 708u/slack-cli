package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// ReactionCmd groups reaction subcommands.
type ReactionCmd struct {
	Add    ReactionAddCmd    `cmd:"" help:"Add reaction"`
	Remove ReactionRemoveCmd `cmd:"" help:"Remove reaction"`
}

// ReactionAddCmd adds an emoji reaction to a message.
type ReactionAddCmd struct {
	Channel   string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	Timestamp string `name:"timestamp" short:"t" required:"" help:"Message timestamp"`
	Emoji     string `name:"emoji" short:"e" required:"" help:"Emoji name (without colons)"`
	Profile   string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the reaction add command.
func (c *ReactionAddCmd) Run() error {
	token, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(token)
	if err := client.AddReaction(c.Channel, c.Timestamp, c.Emoji); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Reaction :%s: added to message in #%s", c.Emoji, c.Channel))
	return nil
}

// ReactionRemoveCmd removes an emoji reaction from a message.
type ReactionRemoveCmd struct {
	Channel   string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	Timestamp string `name:"timestamp" short:"t" required:"" help:"Message timestamp"`
	Emoji     string `name:"emoji" short:"e" required:"" help:"Emoji name (without colons)"`
	Profile   string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the reaction remove command.
func (c *ReactionRemoveCmd) Run() error {
	token, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(token)
	if err := client.RemoveReaction(c.Channel, c.Timestamp, c.Emoji); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Reaction :%s: removed from message in #%s", c.Emoji, c.Channel))
	return nil
}
