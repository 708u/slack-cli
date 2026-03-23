package slack

import (
	"fmt"

	slackgo "github.com/slack-go/slack"
)

// UserGroupOps groups user-group-related API calls.
type UserGroupOps struct {
	api *slackgo.Client
}

func newUserGroupOps(api *slackgo.Client) *UserGroupOps {
	return &UserGroupOps{api: api}
}

// ListUserGroups returns user groups for the workspace.
func (u *UserGroupOps) ListUserGroups(opts ListUserGroupsOptions) ([]UserGroup, error) {
	groups, err := u.api.GetUserGroups(
		slackgo.GetUserGroupsOptionIncludeCount(opts.IncludeCount),
		slackgo.GetUserGroupsOptionIncludeDisabled(opts.IncludeDisabled),
	)
	if err != nil {
		return nil, fmt.Errorf("list user groups: %w", err)
	}

	result := make([]UserGroup, len(groups))
	for i, g := range groups {
		result[i] = userGroupFromSlack(g)
	}
	return result, nil
}

func userGroupFromSlack(g slackgo.UserGroup) UserGroup {
	return UserGroup{
		ID:          g.ID,
		Name:        g.Name,
		Handle:      g.Handle,
		Description: g.Description,
		UserCount:   g.UserCount,
		IsExternal:  g.IsExternal,
		DateCreate:  int64(g.DateCreate),
		DateUpdate:  int64(g.DateUpdate),
	}
}
