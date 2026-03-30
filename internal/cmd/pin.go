package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// PinCmd groups pin subcommands.
type PinCmd struct {
	Add    PinAddCmd    `cmd:"" help:"Pin a message"`
	Remove PinRemoveCmd `cmd:"" help:"Unpin a message"`
	List   PinListCmd   `cmd:"" help:"List pinned items"`
}

// PinAddCmd pins a message in a channel.
type PinAddCmd struct {
	Channel   string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	Timestamp string `name:"timestamp" short:"t" required:"" help:"Message timestamp"`
}

// Run executes the pin add command.
func (c *PinAddCmd) Run(client *slack.Client) error {
	if err := client.AddPin(c.Channel, c.Timestamp); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Pin added to message in #%s", c.Channel))
	return nil
}

// PinRemoveCmd unpins a message from a channel.
type PinRemoveCmd struct {
	Channel   string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	Timestamp string `name:"timestamp" short:"t" required:"" help:"Message timestamp"`
}

// Run executes the pin remove command.
func (c *PinRemoveCmd) Run(client *slack.Client) error {
	if err := client.RemovePin(c.Channel, c.Timestamp); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Pin removed from message in #%s", c.Channel))
	return nil
}

// PinListCmd lists pinned items in a channel.
type PinListCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	Format  string `name:"format" optional:"" default:"table" enum:"table,simple,json" help:"Output format: table, simple, json"`
}

// Run executes the pin list command.
func (c *PinListCmd) Run(client *slack.Client) error {
	items, err := client.ListPins(c.Channel)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Println("No pinned items found")
		return nil
	}

	f := format.ParseFormat(c.Format)
	switch f {
	case format.JSON:
		format.PrintJSON(items)
	case format.Simple:
		for _, item := range items {
			created := format.FormatUnixISO(item.Created)
			text := ""
			ts := ""
			if item.Message != nil {
				text = item.Message.Text
				ts = item.Message.TS
			}
			fmt.Printf("%s %s %s\n", created, ts, text)
		}
	default:
		bold := color.New(color.Bold)
		bold.Printf("%-26s%-20s%-12s%s\n", "Created", "Timestamp", "Creator", "Text")
		fmt.Println(format.Separator(70))
		for _, item := range items {
			created := format.FormatUnixISO(item.Created)
			text := ""
			ts := ""
			if item.Message != nil {
				text = item.Message.Text
				ts = item.Message.TS
			}
			fmt.Printf("%-26s%-20s%-12s%s\n", created, ts, item.CreatedBy, text)
		}
	}
	return nil
}
