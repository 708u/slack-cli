package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/708u/slack-cli/internal/util"
)

// ChannelsCmd lists Slack channels.
type ChannelsCmd struct {
	Type            string `name:"type" help:"Channel type: public,private,im,mpim,all" default:"public"`
	IncludeArchived bool   `name:"include-archived" help:"Include archived channels"`
	Format          string `name:"format" help:"Output format" enum:"table,simple,json" default:"table"`
	Limit           int    `name:"limit" help:"Max channels" default:"100"`
}

// Run executes the channels command.
func (c *ChannelsCmd) Run(client *slack.Client) error {
	types := util.GetChannelTypes(c.Type)
	channels, err := client.ListChannels(slack.ListChannelsOptions{
		Types:           types,
		ExcludeArchived: !c.IncludeArchived,
		Limit:           c.Limit,
	})
	if err != nil {
		// Fallback to users.conversations when conversations.list
		// fails due to missing scopes (e.g. groups:read for private).
		channels, err = client.FetchUserChannels()
		if err != nil {
			return err
		}
	}

	if len(channels) == 0 {
		fmt.Println("No channels found")
		return nil
	}

	infos := make([]format.ChannelInfo, len(channels))
	for i, ch := range channels {
		infos[i] = mapChannelToInfo(ch)
	}

	f := format.ParseFormat(c.Format)
	format.FormatChannelsList(infos, f)
	return nil
}

// mapChannelToInfo converts a slack.Channel to a format.ChannelInfo.
func mapChannelToInfo(ch slack.Channel) format.ChannelInfo {
	purpose := ""
	if ch.Purpose != nil {
		purpose = util.SanitizeTerminalText(ch.Purpose.Value)
	}
	return format.ChannelInfo{
		ID:      ch.ID,
		Name:    util.SanitizeTerminalText(ch.Name),
		Type:    channelTypeLabel(ch),
		Members: ch.NumMembers,
		Created: util.FormatUnixTimestamp(ch.Created),
		Purpose: purpose,
	}
}

// channelTypeLabel returns a human-readable label for the channel type.
func channelTypeLabel(ch slack.Channel) string {
	switch {
	case ch.IsIM:
		return "im"
	case ch.IsMPIM:
		return "mpim"
	case ch.IsPrivate:
		return "private"
	default:
		return "public"
	}
}
