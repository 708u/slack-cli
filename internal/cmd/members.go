package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/config"
	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
)

// MembersCmd lists channel members.
type MembersCmd struct {
	Channel string `name:"channel" short:"c" required:"" help:"Target channel name or ID"`
	Limit   int    `name:"limit" default:"100" help:"Maximum number of members to list"`
	Format  string `name:"format" enum:"table,simple,json" default:"table" help:"Output format"`
	Profile string `name:"profile" optional:"" help:"Use specific workspace profile"`
}

// Run executes the members command.
func (c *MembersCmd) Run() error {
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	client := slack.NewClient(tokens.BotToken, tokens.UserToken)
	result, err := client.GetChannelMembers(c.Channel, &slack.ChannelMembersOptions{
		Limit: c.Limit,
	})
	if err != nil {
		return err
	}

	if len(result.Members) == 0 {
		fmt.Println("No members found")
		return nil
	}

	members := make([]format.MemberInfo, len(result.Members))
	for i, memberID := range result.Members {
		info := format.MemberInfo{ID: memberID}
		user, userErr := client.GetUserInfo(memberID)
		if userErr == nil {
			info.Name = user.Name
			info.RealName = user.RealName
		}
		members[i] = info
	}

	format.FormatMembers(members, format.ParseFormat(c.Format))
	return nil
}
