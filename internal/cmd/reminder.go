package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

// ReminderCmd groups reminder subcommands.
type ReminderCmd struct {
	Add      ReminderAddCmd      `cmd:"" help:"Create reminder"`
	List     ReminderListCmd     `cmd:"" help:"List reminders"`
	Delete   ReminderDeleteCmd   `cmd:"" help:"Delete reminder"`
	Complete ReminderCompleteCmd `cmd:"" help:"Complete reminder"`
}

// ReminderAddCmd creates a new reminder.
type ReminderAddCmd struct {
	Text    string `name:"text" required:"" help:"The content of the reminder"`
	At      string `name:"at" optional:"" help:"Absolute date/time (Unix timestamp or ISO8601)"`
	After   string `name:"after" optional:"" help:"Minutes from now"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Validate ensures exactly one of --at or --after is specified.
func (c *ReminderAddCmd) Validate() error {
	hasAt := c.At != ""
	hasAfter := c.After != ""
	if hasAt == hasAfter {
		return fmt.Errorf("specify exactly one of --at or --after")
	}
	return nil
}

// Run executes the reminder add command.
func (c *ReminderAddCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	postAt, err := resolveTime(c.At, c.After)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	reminder, err := client.AddReminder(c.Text, postAt)
	if err != nil {
		return err
	}

	timeStr := time.Unix(reminder.Time, 0).UTC().Format(time.RFC3339)
	fmt.Println(color.GreenString("Reminder created: %q at %s", reminder.Text, timeStr))
	return nil
}

// ReminderListCmd lists all reminders.
type ReminderListCmd struct {
	Format  string `name:"format" optional:"" default:"table" enum:"table,simple,json" help:"Output format: table, simple, json"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the reminder list command.
func (c *ReminderListCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	reminders, err := client.ListReminders()
	if err != nil {
		return err
	}

	if len(reminders) == 0 {
		fmt.Println("No reminders found")
		return nil
	}

	f := format.ParseFormat(c.Format)
	switch f {
	case format.JSON:
		format.PrintJSON(reminders)
	case format.Simple:
		for _, r := range reminders {
			t := time.Unix(r.Time, 0).UTC().Format(time.RFC3339)
			fmt.Printf("%s %s %s\n", r.ID, t, r.Text)
		}
	default:
		bold := color.New(color.Bold)
		bold.Printf("%-14s%-26s%-10s%s\n", "ID", "Time", "Recurring", "Text")
		fmt.Println(format.Separator(70))
		for _, r := range reminders {
			t := time.Unix(r.Time, 0).UTC().Format(time.RFC3339)
			recurring := "No"
			if r.Recurring {
				recurring = "Yes"
			}
			fmt.Printf("%-14s%-26s%-10s%s\n", r.ID, t, recurring, r.Text)
		}
	}
	return nil
}

// ReminderDeleteCmd deletes a reminder by ID.
type ReminderDeleteCmd struct {
	ID      string `name:"id" required:"" help:"Reminder ID"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the reminder delete command.
func (c *ReminderDeleteCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	if err := client.DeleteReminder(c.ID); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Reminder deleted: %s", c.ID))
	return nil
}

// ReminderCompleteCmd marks a reminder as complete.
type ReminderCompleteCmd struct {
	ID      string `name:"id" required:"" help:"Reminder ID"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the reminder complete command.
func (c *ReminderCompleteCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	if err := client.CompleteReminder(c.ID); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Reminder completed: %s", c.ID))
	return nil
}

// resolveTime converts either an absolute time string or a relative
// minutes string into a Unix timestamp.
//   - If at is set: parse as a Unix timestamp (integer) or ISO8601
//     (RFC3339 / "2006-01-02T15:04:05Z07:00" / "2006-01-02 15:04").
//   - If after is set: parse as integer minutes and add to time.Now().
func resolveTime(at, after string) (int64, error) {
	if at != "" {
		// Try Unix timestamp first.
		if ts, err := strconv.ParseInt(at, 10, 64); err == nil {
			return ts, nil
		}
		// Try RFC3339.
		if t, err := time.Parse(time.RFC3339, at); err == nil {
			return t.Unix(), nil
		}
		// Try "2006-01-02 15:04".
		if t, err := time.Parse("2006-01-02 15:04", at); err == nil {
			return t.Unix(), nil
		}
		return 0, fmt.Errorf("cannot parse --at value %q: use Unix timestamp or ISO8601 format", at)
	}

	minutes, err := strconv.Atoi(after)
	if err != nil {
		return 0, fmt.Errorf("--after must be an integer (minutes): %w", err)
	}
	return time.Now().Add(time.Duration(minutes) * time.Minute).Unix(), nil
}
