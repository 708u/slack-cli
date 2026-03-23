package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/alecthomas/kong"
)

// CLI is the root Kong CLI struct. All subcommands are defined as fields.
type CLI struct {
	Profile string `name:"profile" help:"Workspace profile to use" optional:""`

	Config        ConfigCmd        `cmd:"" help:"Manage Slack CLI configuration"`
	Send          SendCmd          `cmd:"" help:"Send or schedule a message to a Slack channel or DM"`
	Channels      ChannelsCmd      `cmd:"" help:"List Slack channels"`
	History       HistoryCmd       `cmd:"" help:"Get message history from a Slack channel"`
	Unread        UnreadCmd        `cmd:"" help:"Show unread messages across channels"`
	Scheduled     ScheduledCmd     `cmd:"" help:"Manage scheduled messages"`
	Search        SearchCmd        `cmd:"" help:"Search messages in Slack workspace"`
	Edit          EditCmd          `cmd:"" help:"Edit a sent message"`
	Delete        DeleteCmd        `cmd:"" help:"Delete a sent message"`
	Upload        UploadCmd        `cmd:"" help:"Upload a file or snippet to a Slack channel"`
	Reaction      ReactionCmd      `cmd:"" help:"Add or remove emoji reactions on messages"`
	Pin           PinCmd           `cmd:"" help:"Add, remove, or list pinned messages in a channel"`
	Users         UsersCmd         `cmd:"" help:"List, search, and get information about workspace users"`
	Channel       ChannelCmd       `cmd:"" help:"Manage channel topic, purpose, and info"`
	Members       MembersCmd       `cmd:"" help:"List channel members"`
	SendEphemeral SendEphemeralCmd `cmd:"send-ephemeral" help:"Send an ephemeral message visible only to a specific user"`
	Join          JoinCmd          `cmd:"" help:"Join a channel"`
	Leave         LeaveCmd         `cmd:"" help:"Leave a channel"`
	Invite        InviteCmd        `cmd:"" help:"Invite user(s) to a channel"`
	Reminder      ReminderCmd      `cmd:"" help:"Create, list, delete, or complete reminders"`
	Bookmark      BookmarkCmd      `cmd:"" help:"Manage saved items"`
	Canvas        CanvasCmd        `cmd:"" help:"Manage Slack Canvases"`
	Usergroups    UserGroupsCmd    `cmd:"" help:"Manage user groups"`

	Version VersionFlag `name:"version" help:"Print version information"`
}

// ProvideSlackClient creates a *slack.Client from the global --profile flag.
// Kong calls this provider only when a subcommand's Run() accepts *slack.Client.
func (c *CLI) ProvideSlackClient() (*slack.Client, error) {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return nil, err
	}
	return slack.NewClient(tokens.BotToken, tokens.UserToken), nil
}

// VersionFlag is a custom flag type for --version.
type VersionFlag bool

// BeforeApply implements kong.BeforeApply to print version and exit.
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(app.Model.Name, vars["version"])
	app.Exit(0)
	return nil
}
