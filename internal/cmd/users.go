package cmd

import (
	"errors"
	"fmt"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
)

// UsersCmd groups user-related subcommands.
type UsersCmd struct {
	List     UsersListCmd     `cmd:"" help:"List workspace users"`
	Search   UsersSearchCmd   `cmd:"" help:"Search users by name"`
	Info     UsersInfoCmd     `cmd:"" help:"Get user info"`
	Lookup   UsersLookupCmd   `cmd:"" help:"Look up user by email"`
	Presence UsersPresenceCmd `cmd:"" help:"Check user presence"`
}

// UsersListCmd lists workspace users.
type UsersListCmd struct {
	Limit   int    `name:"limit" default:"100" help:"Maximum number of users to list"`
	Format  string `name:"format" enum:"table,simple,json" default:"table" help:"Output format"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the users list command.
func (c *UsersListCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	users, err := client.ListUsers(c.Limit)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("No users found")
		return nil
	}

	infos := make([]format.UserInfo, len(users))
	for i, u := range users {
		infos[i] = toUserInfo(u)
	}

	format.FormatUserList(infos, format.ParseFormat(c.Format))
	return nil
}

// UsersInfoCmd displays detailed information about a user.
type UsersInfoCmd struct {
	ID      string `name:"id" required:"" help:"User ID"`
	Format  string `name:"format" enum:"table,simple,json" default:"table" help:"Output format"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the users info command.
func (c *UsersInfoCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	user, err := client.GetUserInfo(c.ID)
	if err != nil {
		return err
	}

	format.FormatUserInfo(toUserDetailInfo(user), format.ParseFormat(c.Format))
	return nil
}

// UsersLookupCmd looks up a user by email address.
type UsersLookupCmd struct {
	Email   string `name:"email" required:"" help:"Email address to look up"`
	Format  string `name:"format" enum:"table,simple,json" default:"table" help:"Output format"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the users lookup command.
func (c *UsersLookupCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	user, err := client.LookupUserByEmail(c.Email)
	if err != nil {
		return err
	}

	format.FormatUserInfo(toUserDetailInfo(user), format.ParseFormat(c.Format))
	return nil
}

// UsersPresenceCmd checks the presence status of a user.
type UsersPresenceCmd struct {
	ID      string `name:"id" optional:"" help:"User ID"`
	Name    string `name:"name" optional:"" help:"Username (e.g. @username)"`
	Format  string `name:"format" enum:"table,simple,json" default:"table" help:"Output format"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the users presence command.
func (c *UsersPresenceCmd) Run() error {
	if c.ID == "" && c.Name == "" {
		return errors.New("you must specify either --id or --name")
	}
	if c.ID != "" && c.Name != "" {
		return errors.New("cannot use both --id and --name")
	}

	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)

	userID := c.ID
	if c.Name != "" {
		userID, err = client.ResolveUserIDByName(c.Name)
		if err != nil {
			return err
		}
	}

	presence, err := client.GetUserPresence(userID)
	if err != nil {
		return err
	}

	format.FormatPresence(userID, presence.Presence, format.ParseFormat(c.Format))
	return nil
}

// UsersSearchCmd searches users by real name or display name.
type UsersSearchCmd struct {
	Query   string `arg:"" help:"Search query (matches real name and display name)"`
	Limit   int    `name:"limit" default:"50" help:"Maximum number of results"`
	Format  string `name:"format" enum:"table,simple,json" default:"table" help:"Output format"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the users search command.
func (c *UsersSearchCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	users, err := client.SearchUsers(c.Query, c.Limit)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Printf("No users found for query %q\n", c.Query)
		return nil
	}

	infos := make([]format.UserInfo, len(users))
	for i, u := range users {
		infos[i] = toUserInfo(u)
	}

	format.FormatUserList(infos, format.ParseFormat(c.Format))
	return nil
}

func toUserInfo(u slack.SlackUser) format.UserInfo {
	info := format.UserInfo{
		ID:       u.ID,
		Name:     u.Name,
		RealName: u.RealName,
		IsBot:    u.IsBot,
		IsAdmin:  u.IsAdmin,
	}
	if u.Profile != nil {
		info.DisplayName = u.Profile.DisplayName
	}
	return info
}

func toUserDetailInfo(u *slack.SlackUser) format.UserDetailInfo {
	info := format.UserDetailInfo{
		ID:       u.ID,
		Name:     u.Name,
		RealName: u.RealName,
		TZ:       u.TZ,
		TZLabel:  u.TZLabel,
		IsAdmin:  u.IsAdmin,
		IsBot:    u.IsBot,
		Deleted:  u.Deleted,
	}
	if u.Profile != nil {
		info.Email = u.Profile.Email
		info.DisplayName = u.Profile.DisplayName
		info.Title = u.Profile.Title
		info.StatusText = u.Profile.StatusText
		info.StatusEmoji = u.Profile.StatusEmoji
	}
	return info
}
