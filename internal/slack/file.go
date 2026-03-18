package slack

import (
	"fmt"
	"os"
	"path/filepath"

	slackgo "github.com/slack-go/slack"
)

// FileOps groups file upload API calls.
type FileOps struct {
	api        *slackgo.Client
	channelOps *ChannelOps
}

func newFileOps(api *slackgo.Client, channelOps *ChannelOps) *FileOps {
	return &FileOps{
		api:        api,
		channelOps: channelOps,
	}
}

// UploadFile uploads a file or code snippet to the specified channel.
// Either FilePath or Content must be provided in opts. The channel
// field accepts a name or ID.
func (f *FileOps) UploadFile(opts UploadFileOptions) error {
	channelID, err := f.channelOps.ResolveChannelID(opts.Channel)
	if err != nil {
		return err
	}

	params := slackgo.UploadFileParameters{
		Channel:         channelID,
		Title:           opts.Title,
		InitialComment:  opts.InitialComment,
		SnippetType:     opts.SnippetType,
		ThreadTimestamp: opts.ThreadTS,
	}

	if opts.FilePath != "" {
		params.File = opts.FilePath
		if opts.Filename != "" {
			params.Filename = opts.Filename
		} else {
			params.Filename = filepath.Base(opts.FilePath)
		}

		info, err := os.Stat(opts.FilePath)
		if err != nil {
			return fmt.Errorf("upload file: stat: %w", err)
		}
		params.FileSize = int(info.Size())
	} else if opts.Content != "" {
		params.Content = opts.Content
		params.Filename = opts.Filename
		if params.Filename == "" {
			params.Filename = "snippet.txt"
		}
		params.FileSize = len(opts.Content)
	} else {
		return fmt.Errorf("upload file: either FilePath or Content must be provided")
	}

	if _, err := f.api.UploadFile(params); err != nil {
		return fmt.Errorf("upload file: %w", err)
	}
	return nil
}
