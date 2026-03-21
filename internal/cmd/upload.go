package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/708u/slack-cli/internal/util"
)

// UploadCmd uploads a file or snippet to a Slack channel or DM.
type UploadCmd struct {
	Channel  string `name:"channel" short:"c" optional:"" help:"Channel name or ID"`
	User     string `name:"user" short:"u" optional:"" help:"Upload to DM by username or user ID"`
	File     string `name:"file" short:"f" optional:"" help:"File path to upload"`
	Content  string `name:"content" optional:"" help:"Text content to upload as snippet"`
	Filename string `name:"filename" optional:"" help:"Override filename"`
	Title    string `name:"title" optional:"" help:"File title"`
	Message  string `name:"message" short:"m" optional:"" help:"Initial comment with the file"`
	Filetype string `name:"filetype" optional:"" help:"Snippet type (e.g. python, javascript, csv)"`
	Thread   string `name:"thread" short:"t" optional:"" help:"Thread timestamp to upload as reply"`
	Profile  string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Validate checks that exactly one of --file or --content is provided
// and that exactly one target (--channel or --user) is given.
func (c *UploadCmd) Validate() error {
	if c.Channel == "" && c.User == "" {
		return errors.New("you must specify --channel or --user")
	}
	if c.Channel != "" && c.User != "" {
		return errors.New("cannot use --channel and --user together")
	}
	if c.File == "" && c.Content == "" {
		return errors.New("you must specify either --file or --content")
	}
	if c.File != "" && c.Content != "" {
		return errors.New("cannot use both --file and --content")
	}
	return nil
}

// Run executes the upload command.
func (c *UploadCmd) Run() error {
	if c.File != "" {
		if _, err := os.Stat(c.File); err != nil {
			return util.NewFileError(fmt.Sprintf("file not found: %s", c.File))
		}
	}

	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)

	channelID, displayName, err := c.resolveTarget(client)
	if err != nil {
		return err
	}

	opts := slack.UploadFileOptions{
		Channel:        channelID,
		FilePath:       c.File,
		Content:        c.Content,
		Filename:       c.Filename,
		Title:          c.Title,
		InitialComment: c.Message,
		SnippetType:    c.Filetype,
		ThreadTS:       c.Thread,
	}

	if err := client.UploadFile(opts); err != nil {
		return err
	}

	fmt.Printf("File uploaded successfully to %s\n", displayName)
	return nil
}

func (c *UploadCmd) resolveTarget(client *slack.Client) (channelID, displayName string, err error) {
	if c.User != "" {
		userID, err := client.ResolveUserID(c.User)
		if err != nil {
			return "", "", err
		}
		ch, err := client.OpenDMChannel(userID)
		if err != nil {
			return "", "", err
		}
		return ch, "@" + strings.TrimPrefix(c.User, "@"), nil
	}
	ch, err := client.ResolveChannelID(c.Channel)
	if err != nil {
		return "", "", err
	}
	return ch, "#" + c.Channel, nil
}
