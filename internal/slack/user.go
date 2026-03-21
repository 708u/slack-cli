package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/708u/slack-cli/internal/util"
	slackgo "github.com/slack-go/slack"
)

func newUserOps(api *slackgo.Client) *UserOps {
	return &UserOps{api: api}
}

// ListUsers returns workspace members using cursor-based pagination.
// If limit <= 0, all users are returned.
func (u *UserOps) ListUsers(limit int) ([]SlackUser, error) {
	if limit < 0 {
		return nil, nil
	}

	var users []SlackUser

	p := u.api.GetUsersPaginated(slackgo.GetUsersOptionLimit(defaultFetchLimit))
	for {
		var err error
		p, err = p.Next(context.Background())
		if p.Done(err) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}

		for _, su := range p.Users {
			users = append(users, userFromSlack(su))
			if limit > 0 && len(users) >= limit {
				return users, nil
			}
		}
	}

	return users, nil
}

// GetUserInfo returns profile information for a single user.
func (u *UserOps) GetUserInfo(userID string) (*SlackUser, error) {
	su, err := u.api.GetUserInfo(userID)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}
	user := userFromSlack(*su)
	return &user, nil
}

// LookupByEmail finds a user by their email address.
func (u *UserOps) LookupByEmail(email string) (*SlackUser, error) {
	su, err := u.api.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("lookup by email: %w", err)
	}
	user := userFromSlack(*su)
	return &user, nil
}

// OpenDMChannel opens a direct-message conversation with the given user
// and returns the channel ID.
func (u *UserOps) OpenDMChannel(userID string) (string, error) {
	ch, _, _, err := u.api.OpenConversation(&slackgo.OpenConversationParameters{
		Users: []string{userID},
	})
	if err != nil {
		return "", fmt.Errorf("open DM channel: %w", err)
	}
	return ch.ID, nil
}

// GetPresence returns the online/away presence status of a user.
func (u *UserOps) GetPresence(userID string) (*UserPresence, error) {
	p, err := u.api.GetUserPresence(userID)
	if err != nil {
		return nil, fmt.Errorf("get presence: %w", err)
	}
	return &UserPresence{Presence: p.Presence}, nil
}

// ResolveUserIDByName finds a user ID by @username handle. It strips a
// leading '@' and searches through the full users.list with
// case-insensitive matching.
func (u *UserOps) ResolveUserIDByName(username string) (string, error) {
	name := strings.TrimPrefix(username, "@")
	nameLower := strings.ToLower(name)

	p := u.api.GetUsersPaginated(slackgo.GetUsersOptionLimit(defaultFetchLimit))
	for {
		var err error
		p, err = p.Next(context.Background())
		if p.Done(err) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("resolve user: %w", err)
		}

		for _, su := range p.Users {
			if strings.ToLower(su.Name) == nameLower {
				return su.ID, nil
			}
		}
	}

	return "", fmt.Errorf("user '%s' not found", name)
}

// ResolveUserID accepts a username or user ID and returns the user ID.
// If the input already looks like a user ID (U/W prefix), it is returned
// immediately without an API call. Otherwise it falls through to
// ResolveUserIDByName.
func (u *UserOps) ResolveUserID(nameOrID string) (string, error) {
	if util.IsUserID(nameOrID) {
		return nameOrID, nil
	}
	return u.ResolveUserIDByName(nameOrID)
}

// userFromSlack converts a slack-go User to the internal SlackUser type.
func userFromSlack(su slackgo.User) SlackUser {
	return SlackUser{
		ID:       su.ID,
		Name:     su.Name,
		RealName: su.RealName,
		Profile: &UserProfile{
			Email:       su.Profile.Email,
			DisplayName: su.Profile.DisplayName,
			Title:       su.Profile.Title,
			StatusText:  su.Profile.StatusText,
			StatusEmoji: su.Profile.StatusEmoji,
		},
		TZ:      su.TZ,
		TZLabel: su.TZLabel,
		IsAdmin: su.IsAdmin,
		IsBot:   su.IsBot,
		Deleted: su.Deleted,
	}
}
