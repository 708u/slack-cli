package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/708u/slack-cli/internal/slack"
	"github.com/fatih/color"
)

var threadTSPattern = regexp.MustCompile(`^\d{10}\.\d{6}$`)

// SendCmd sends or schedules a message to a Slack channel or DM.
type SendCmd struct {
	Channel string `name:"channel" short:"c" help:"Target channel name or ID" optional:""`
	User    string `name:"user" help:"Send DM by username" optional:""`
	UserID  string `name:"user-id" help:"Send DM by user ID" optional:""`
	Email   string `name:"email" help:"Send DM by email address" optional:""`
	Message string `name:"message" short:"m" help:"Message text" optional:""`
	File    string `name:"file" short:"f" help:"File containing message content" optional:""`
	Thread  string `name:"thread" short:"t" help:"Thread timestamp to reply to" optional:""`
	At      string `name:"at" help:"Schedule time (Unix timestamp in seconds or ISO 8601)" optional:""`
	After   string `name:"after" help:"Schedule after N minutes" optional:""`
}

// Run executes the send command.
func (c *SendCmd) Run(client *slack.Client) error {
	if err := c.validate(); err != nil {
		return err
	}

	messageContent, err := c.resolveMessage()
	if err != nil {
		return err
	}

	postAt, err := resolvePostAt(c.At, c.After)
	if err != nil {
		return err
	}

	targetChannel, targetLabel, err := c.resolveTarget(client)
	if err != nil {
		return err
	}

	if postAt > 0 {
		if err := client.ScheduleMessage(targetChannel, messageContent, postAt, c.Thread); err != nil {
			return err
		}
		postAtISO := time.Unix(postAt, 0).UTC().Format(time.RFC3339)
		fmt.Println(color.GreenString("Message scheduled to %s at %s", targetLabel, postAtISO))
		return nil
	}

	if err := client.SendMessage(targetChannel, messageContent, c.Thread); err != nil {
		return err
	}

	if c.User != "" || c.UserID != "" || c.Email != "" {
		fmt.Println(color.GreenString("DM sent to %s", targetLabel))
	} else {
		fmt.Println(color.GreenString("Message sent successfully to #%s", c.Channel))
	}
	return nil
}

func (c *SendCmd) validate() error {
	// Exactly one target is required.
	targets := 0
	if c.Channel != "" {
		targets++
	}
	if c.User != "" {
		targets++
	}
	if c.UserID != "" {
		targets++
	}
	if c.Email != "" {
		targets++
	}

	if targets == 0 {
		return fmt.Errorf("you must specify one of: --channel, --user, --user-id, or --email")
	}
	if targets > 1 {
		return fmt.Errorf("--channel, --user, --user-id, and --email are mutually exclusive")
	}

	// Exactly one message source is required.
	if c.Message == "" && c.File == "" {
		return fmt.Errorf("you must specify either --message or --file")
	}
	if c.Message != "" && c.File != "" {
		return fmt.Errorf("cannot use both --message and --file")
	}

	// Thread timestamp format.
	if c.Thread != "" && !threadTSPattern.MatchString(c.Thread) {
		return fmt.Errorf("invalid thread timestamp format")
	}

	// Schedule options.
	if c.At != "" && c.After != "" {
		return fmt.Errorf("cannot use both --at and --after")
	}

	return nil
}

func (c *SendCmd) resolveMessage() (string, error) {
	if c.File != "" {
		data, err := os.ReadFile(c.File)
		if err != nil {
			return "", fmt.Errorf("error reading file %s: %w", c.File, err)
		}
		return string(data), nil
	}
	return c.Message, nil
}

func (c *SendCmd) resolveTarget(client *slack.Client) (channel, label string, err error) {
	if c.User != "" {
		userID, err := client.ResolveUserIDByName(c.User)
		if err != nil {
			return "", "", err
		}
		ch, err := client.OpenDMChannel(userID)
		if err != nil {
			return "", "", err
		}
		return ch, "@" + strings.TrimPrefix(c.User, "@"), nil
	}

	if c.UserID != "" {
		ch, err := client.OpenDMChannel(c.UserID)
		if err != nil {
			return "", "", err
		}
		return ch, c.UserID, nil
	}

	if c.Email != "" {
		user, err := client.LookupUserByEmail(c.Email)
		if err != nil {
			return "", "", err
		}
		ch, err := client.OpenDMChannel(user.ID)
		if err != nil {
			return "", "", err
		}
		return ch, c.Email, nil
	}

	// Resolve channel name to ID so that private channels work
	// even when the bot token cannot see them.
	channelID, err := client.ResolveChannelID(c.Channel)
	if err != nil {
		return "", "", err
	}
	return channelID, "#" + c.Channel, nil
}

// resolvePostAt parses the --at or --after flags into a Unix timestamp.
// Returns 0 when no scheduling is requested.
func resolvePostAt(at, after string) (int64, error) {
	if at != "" {
		return parseScheduledTimestamp(at)
	}

	if after == "" {
		return 0, nil
	}

	trimmed := strings.TrimSpace(after)
	minutes, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil || minutes <= 0 {
		return 0, fmt.Errorf("--after must be a positive integer (minutes)")
	}

	return time.Now().Unix() + minutes*60, nil
}

// parseScheduledTimestamp parses a Unix timestamp (seconds) or ISO 8601
// string into a Unix epoch in seconds.
func parseScheduledTimestamp(value string) (int64, error) {
	trimmed := strings.TrimSpace(value)

	// Try pure numeric (Unix timestamp in seconds).
	if ts, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return ts, nil
	}

	// Try ISO 8601.
	t, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid schedule time format. Use Unix timestamp (seconds) or ISO 8601 date-time",
		)
	}
	return t.Unix(), nil
}
