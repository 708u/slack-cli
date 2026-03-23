package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
)

// UserGroupsCmd groups user-group subcommands.
type UserGroupsCmd struct {
	List UserGroupsListCmd `cmd:"" help:"List user groups"`
}

// UserGroupsListCmd lists user groups in the workspace.
type UserGroupsListCmd struct {
	IncludeDisabled bool   `name:"include-disabled" default:"false" help:"Include disabled user groups"`
	Format          string `name:"format" enum:"table,simple,json" default:"table" help:"Output format"`
}

// Run executes the usergroups list command.
func (c *UserGroupsListCmd) Run(client *slack.Client) error {
	groups, err := client.ListUserGroups(slack.ListUserGroupsOptions{
		IncludeCount:    true,
		IncludeDisabled: c.IncludeDisabled,
	})
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		fmt.Println("No user groups found")
		return nil
	}

	infos := make([]format.UserGroupInfo, len(groups))
	for i, g := range groups {
		infos[i] = format.UserGroupInfo{
			ID:          g.ID,
			Name:        g.Name,
			Handle:      g.Handle,
			Description: g.Description,
			UserCount:   g.UserCount,
			IsExternal:  g.IsExternal,
		}
	}

	format.FormatUserGroupList(infos, format.ParseFormat(c.Format))
	return nil
}
