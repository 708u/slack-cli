package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/708u/slack-cli/internal/util"
)

// ChannelCmd groups channel management subcommands (info, set-topic,
// set-purpose). This is distinct from ChannelsCmd which lists channels.
type ChannelCmd struct {
	Info       ChannelInfoCmd       `cmd:"" help:"Display channel details"`
	SetTopic   ChannelSetTopicCmd   `cmd:"set-topic" help:"Set channel topic"`
	SetPurpose ChannelSetPurposeCmd `cmd:"set-purpose" help:"Set channel purpose"`
}

// ChannelInfoCmd displays detailed information about a channel.
type ChannelInfoCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Target channel name or ID"`
	Format  string `name:"format" enum:"table,simple,json" default:"table" help:"Output format"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the channel info command.
func (c *ChannelInfoCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	detail, err := client.GetChannelDetail(c.Channel)
	if err != nil {
		return err
	}

	format.FormatChannelInfo(toChannelDetailInfo(detail), format.ParseFormat(c.Format))
	return nil
}

// ChannelSetTopicCmd sets the topic of a channel.
type ChannelSetTopicCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Target channel name or ID"`
	Topic   string `name:"topic" required:"" help:"New topic text"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the channel set-topic command.
func (c *ChannelSetTopicCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	if err := client.SetTopic(c.Channel, c.Topic); err != nil {
		return err
	}

	fmt.Printf("Topic updated for #%s\n", c.Channel)
	return nil
}

// ChannelSetPurposeCmd sets the purpose of a channel.
type ChannelSetPurposeCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Target channel name or ID"`
	Purpose string `name:"purpose" required:"" help:"New purpose text"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the channel set-purpose command.
func (c *ChannelSetPurposeCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	if err := client.SetPurpose(c.Channel, c.Purpose); err != nil {
		return err
	}

	fmt.Printf("Purpose updated for #%s\n", c.Channel)
	return nil
}

func toChannelDetailInfo(d *slack.ChannelDetail) format.ChannelDetailInfo {
	info := format.ChannelDetailInfo{
		ID:         d.ID,
		Name:       d.Name,
		IsPrivate:  d.IsPrivate,
		IsArchived: d.IsArchived,
		Created:    util.FormatUnixTimestamp(d.Created),
		NumMembers: d.NumMembers,
	}
	if d.Topic != nil {
		info.Topic = d.Topic.Value
	}
	if d.Purpose != nil {
		info.Purpose = d.Purpose.Value
	}
	return info
}
