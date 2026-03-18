package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/708u/slack-cli/internal/util"
)

// UploadCmd uploads a file or snippet to a Slack channel.
type UploadCmd struct {
	Channel  string `name:"channel" short:"c" required:"" help:"Channel name or ID"`
	File     string `name:"file" short:"f" optional:"" help:"File path to upload"`
	Content  string `name:"content" optional:"" help:"Text content to upload as snippet"`
	Filename string `name:"filename" optional:"" help:"Override filename"`
	Title    string `name:"title" optional:"" help:"File title"`
	Message  string `name:"message" short:"m" optional:"" help:"Initial comment with the file"`
	Filetype string `name:"filetype" optional:"" help:"Snippet type (e.g. python, javascript, csv)"`
	Thread   string `name:"thread" short:"t" optional:"" help:"Thread timestamp to upload as reply"`
	Profile  string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Validate checks that exactly one of --file or --content is provided.
func (c *UploadCmd) Validate() error {
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
	opts := slack.UploadFileOptions{
		Channel:        c.Channel,
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

	fmt.Printf("File uploaded successfully to #%s\n", c.Channel)
	return nil
}
