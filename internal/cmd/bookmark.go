package cmd

import (
	"fmt"
	"time"

	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// BookmarkCmd groups bookmark (saved items) subcommands.
type BookmarkCmd struct {
	Add    BookmarkAddCmd    `cmd:"" help:"Save a message"`
	List   BookmarkListCmd   `cmd:"" help:"List saved items"`
	Remove BookmarkRemoveCmd `cmd:"" help:"Remove saved item"`
}

// BookmarkAddCmd saves a message for later.
type BookmarkAddCmd struct {
	Channel   string `name:"channel" short:"c" required:"" help:"Channel ID"`
	Timestamp string `name:"ts" required:"" help:"Message timestamp"`
}

// Run executes the bookmark add command.
func (c *BookmarkAddCmd) Run(client *slack.Client) error {
	if err := client.AddStar(c.Channel, c.Timestamp); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Saved message %s in %s", c.Timestamp, c.Channel))
	return nil
}

// BookmarkListCmd lists saved items.
type BookmarkListCmd struct {
	Limit  int    `name:"limit" optional:"" default:"100" help:"Number of items to display"`
	Format string `name:"format" optional:"" default:"table" enum:"table,simple,json" help:"Output format: table, simple, json"`
}

// Run executes the bookmark list command.
func (c *BookmarkListCmd) Run(client *slack.Client) error {
	result, err := client.ListStars(c.Limit)
	if err != nil {
		return err
	}

	if len(result.Items) == 0 {
		fmt.Println("No saved items found")
		return nil
	}

	f := format.ParseFormat(c.Format)
	switch f {
	case format.JSON:
		format.PrintJSON(result.Items)
	case format.Simple:
		for _, item := range result.Items {
			created := ""
			if item.DateCreate != 0 {
				created = time.Unix(item.DateCreate, 0).UTC().Format(time.RFC3339)
			}
			fmt.Printf("%s %s %s %s\n", created, item.Channel, item.Message.TS, item.Message.Text)
		}
	default:
		bold := color.New(color.Bold)
		bold.Printf("%-26s%-14s%-20s%s\n", "Created", "Channel", "Timestamp", "Text")
		fmt.Println(format.Separator(70))
		for _, item := range result.Items {
			created := ""
			if item.DateCreate != 0 {
				created = time.Unix(item.DateCreate, 0).UTC().Format(time.RFC3339)
			}
			fmt.Printf("%-26s%-14s%-20s%s\n", created, item.Channel, item.Message.TS, item.Message.Text)
		}
	}
	return nil
}

// BookmarkRemoveCmd removes a saved item.
type BookmarkRemoveCmd struct {
	Channel   string `name:"channel" short:"c" required:"" help:"Channel ID"`
	Timestamp string `name:"ts" required:"" help:"Message timestamp"`
}

// Run executes the bookmark remove command.
func (c *BookmarkRemoveCmd) Run(client *slack.Client) error {
	if err := client.RemoveStar(c.Channel, c.Timestamp); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Removed saved item %s from %s", c.Timestamp, c.Channel))
	return nil
}
