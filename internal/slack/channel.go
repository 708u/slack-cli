package slack

import (
	"fmt"
	"strings"
	"sync"

	slackgo "github.com/slack-go/slack"
	"golang.org/x/sync/errgroup"

	"github.com/708u/slack-cli/internal/util"
)

const (
	defaultFetchLimit            = 200
	defaultMembersLimit          = 100
	unreadScanConcurrentRequests = 15
)

// ChannelOps groups channel-related API calls. It caches the channel
// list for repeated name-to-ID resolution.
type ChannelOps struct {
	api          *slackgo.Client
	channelCache []Channel
	mu           sync.Mutex
}

func newChannelOps(api *slackgo.Client) *ChannelOps {
	return &ChannelOps{api: api}
}

// ListChannels returns channels matching the given options, paginating
// through the full result set using cursor-based pagination.
func (c *ChannelOps) ListChannels(opts ListChannelsOptions) ([]Channel, error) {
	var channels []Channel

	types := strings.Split(opts.Types, ",")
	params := &slackgo.GetConversationsParameters{
		Types:           types,
		ExcludeArchived: opts.ExcludeArchived,
		Limit:           opts.Limit,
	}

	for {
		slackChannels, nextCursor, err := c.api.GetConversations(params)
		if err != nil {
			return nil, fmt.Errorf("list channels: %w", err)
		}

		for _, sc := range slackChannels {
			channels = append(channels, channelFromSlack(sc))
		}

		if nextCursor == "" {
			break
		}
		params.Cursor = nextCursor
	}

	return channels, nil
}

// GetChannelDetail retrieves detailed information about a single
// channel, including the member count.
func (c *ChannelOps) GetChannelDetail(nameOrID string) (*ChannelDetail, error) {
	channelID, err := c.ResolveChannelID(nameOrID)
	if err != nil {
		return nil, err
	}

	sc, err := c.api.GetConversationInfo(&slackgo.GetConversationInfoInput{
		ChannelID:         channelID,
		IncludeNumMembers: true,
	})
	if err != nil {
		return nil, fmt.Errorf("get channel detail: %w", err)
	}

	detail := channelDetailFromSlack(*sc)
	return &detail, nil
}

// SetTopic sets the topic of the specified channel.
func (c *ChannelOps) SetTopic(nameOrID, topic string) error {
	channelID, err := c.ResolveChannelID(nameOrID)
	if err != nil {
		return err
	}

	if _, err := c.api.SetTopicOfConversation(channelID, topic); err != nil {
		return fmt.Errorf("set topic: %w", err)
	}
	return nil
}

// SetPurpose sets the purpose of the specified channel.
func (c *ChannelOps) SetPurpose(nameOrID, purpose string) error {
	channelID, err := c.ResolveChannelID(nameOrID)
	if err != nil {
		return err
	}

	if _, err := c.api.SetPurposeOfConversation(channelID, purpose); err != nil {
		return fmt.Errorf("set purpose: %w", err)
	}
	return nil
}

// JoinChannel joins the specified channel.
func (c *ChannelOps) JoinChannel(nameOrID string) error {
	channelID, err := c.ResolveChannelID(nameOrID)
	if err != nil {
		return err
	}

	if _, _, _, err := c.api.JoinConversation(channelID); err != nil {
		return fmt.Errorf("join channel: %w", err)
	}
	return nil
}

// LeaveChannel leaves the specified channel.
func (c *ChannelOps) LeaveChannel(nameOrID string) error {
	channelID, err := c.ResolveChannelID(nameOrID)
	if err != nil {
		return err
	}

	if _, err := c.api.LeaveConversation(channelID); err != nil {
		return fmt.Errorf("leave channel: %w", err)
	}
	return nil
}

// InviteToChannel invites the given users to the specified channel. The
// force parameter is accepted for API parity but slack-go's
// InviteUsersToConversation does not support it directly.
func (c *ChannelOps) InviteToChannel(nameOrID string, userIDs []string, force bool) error {
	channelID, err := c.ResolveChannelID(nameOrID)
	if err != nil {
		return err
	}

	if _, err := c.api.InviteUsersToConversation(channelID, userIDs...); err != nil {
		return fmt.Errorf("invite to channel: %w", err)
	}
	return nil
}

// GetChannelMembers returns a page of member user IDs for the specified
// channel.
func (c *ChannelOps) GetChannelMembers(nameOrID string, opts ChannelMembersOptions) (*ChannelMembersResult, error) {
	channelID, err := c.ResolveChannelID(nameOrID)
	if err != nil {
		return nil, err
	}

	limit := opts.Limit
	if limit == 0 {
		limit = defaultMembersLimit
	}

	members, nextCursor, err := c.api.GetUsersInConversation(&slackgo.GetUsersInConversationParameters{
		ChannelID: channelID,
		Cursor:    opts.Cursor,
		Limit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("get channel members: %w", err)
	}

	return &ChannelMembersResult{
		Members:    members,
		NextCursor: nextCursor,
	}, nil
}

// ResolveChannelID converts a channel name or ID to a channel ID. If
// nameOrID already matches the channel ID pattern ([CDG][A-Z0-9]{8,}),
// it is returned as-is. Otherwise the cached channel list is searched
// (case-insensitive, with or without the '#' prefix).
func (c *ChannelOps) ResolveChannelID(nameOrID string) (string, error) {
	if isChannelID(nameOrID) {
		return nameOrID, nil
	}

	channels, err := c.getChannelLookupCache()
	if err != nil {
		return "", err
	}

	ch := findChannel(nameOrID, channels)
	if ch == nil {
		return "", resolveChannelError(nameOrID, channels)
	}
	return ch.ID, nil
}

// ListUnreadChannels returns the user's channels that have unread
// messages. It fetches all user conversations, then concurrently checks
// each for unread status.
func (c *ChannelOps) ListUnreadChannels() ([]Channel, error) {
	channels, err := c.FetchUserChannels()
	if err != nil {
		return nil, err
	}

	sem := make(chan struct{}, unreadScanConcurrentRequests)
	var g errgroup.Group
	var mu sync.Mutex
	var unread []Channel

	for _, ch := range channels {
		ch := ch
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := c.getChannelUnreadInfo(ch)
			if err != nil {
				// Non-fatal: skip channels that fail (rate limit, etc.)
				return nil
			}
			if result != nil {
				mu.Lock()
				unread = append(unread, *result)
				mu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return unread, nil
}

// FetchUserChannels returns all conversations the authenticated user
// belongs to (public, private, IM, MPIM), excluding archived channels.
func (c *ChannelOps) FetchUserChannels() ([]Channel, error) {
	var channels []Channel

	params := &slackgo.GetConversationsForUserParameters{
		Types:           []string{"public_channel", "private_channel", "im", "mpim"},
		ExcludeArchived: true,
		Limit:           defaultFetchLimit,
	}

	for {
		slackChannels, nextCursor, err := c.api.GetConversationsForUser(params)
		if err != nil {
			return nil, fmt.Errorf("fetch user channels: %w", err)
		}

		for _, sc := range slackChannels {
			channels = append(channels, channelFromSlack(sc))
		}

		if nextCursor == "" {
			break
		}
		params.Cursor = nextCursor
	}

	return channels, nil
}

// getChannelLookupCache returns the cached channel list, fetching it on
// first call. The cache is protected by a mutex for concurrent access.
func (c *ChannelOps) getChannelLookupCache() ([]Channel, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channelCache != nil {
		return c.channelCache, nil
	}

	// Use users.conversations instead of conversations.list so that
	// private channels the bot has joined are visible without the
	// groups:read scope.
	channels, err := c.FetchUserChannels()
	if err != nil {
		return nil, err
	}

	c.channelCache = channels
	return c.channelCache, nil
}

// getChannelUnreadInfo checks whether a channel has unread messages and
// returns an enriched Channel if so. Returns nil if the channel has no
// unread messages.
func (c *ChannelOps) getChannelUnreadInfo(ch Channel) (*Channel, error) {
	hasUnreadCount := ch.UnreadCount > 0 || ch.UnreadCountDisp > 0
	needsInfo := !hasUnreadCount || (ch.Name == "" && !ch.IsIM && !ch.IsMPIM)

	var info *slackgo.Channel
	if needsInfo {
		var err error
		info, err = c.api.GetConversationInfo(&slackgo.GetConversationInfoInput{
			ChannelID:         ch.ID,
			IncludeNumMembers: false,
		})
		if err != nil {
			return nil, fmt.Errorf("get channel info: %w", err)
		}
	}

	unreadCount := ch.UnreadCountDisp
	if unreadCount == 0 {
		unreadCount = ch.UnreadCount
	}
	if unreadCount == 0 && info != nil {
		unreadCount = info.UnreadCountDisplay
		if unreadCount == 0 {
			unreadCount = info.UnreadCount
		}
	}

	if unreadCount <= 0 {
		return nil, nil
	}

	result := buildUnreadChannel(ch, info, unreadCount)
	return &result, nil
}

// isChannelID reports whether s matches the Slack channel ID pattern.
func isChannelID(s string) bool {
	return util.IsChannelID(s)
}

// findChannel searches for a channel by name in the given list. It
// tries exact match, match without '#' prefix, case-insensitive match,
// and normalized name match.
func findChannel(name string, channels []Channel) *Channel {
	stripped := strings.TrimPrefix(name, "#")
	lower := strings.ToLower(name)

	for i := range channels {
		ch := &channels[i]
		switch {
		case ch.Name == name:
			return ch
		case ch.Name == stripped:
			return ch
		case strings.EqualFold(ch.Name, lower):
			return ch
		case ch.NameNormalized == name:
			return ch
		}
	}
	return nil
}

// getSimilarChannels returns up to limit channel names that contain
// nameOrID as a substring (case-insensitive).
func getSimilarChannels(nameOrID string, channels []Channel, limit int) []string {
	lower := strings.ToLower(nameOrID)
	var similar []string

	for _, ch := range channels {
		if strings.Contains(strings.ToLower(ch.Name), lower) {
			similar = append(similar, ch.Name)
			if len(similar) >= limit {
				break
			}
		}
	}
	return similar
}

// resolveChannelError builds an error with channel name suggestions
// when a channel cannot be found.
func resolveChannelError(nameOrID string, channels []Channel) error {
	sanitized := util.SanitizeTerminalText(nameOrID)
	similar := getSimilarChannels(nameOrID, channels, 5)

	if len(similar) > 0 {
		sanitizedSuggestions := make([]string, len(similar))
		for i, s := range similar {
			sanitizedSuggestions[i] = util.SanitizeTerminalText(s)
		}
		return fmt.Errorf(
			"channel '%s' not found. Did you mean one of these? %s",
			sanitized, strings.Join(sanitizedSuggestions, ", "),
		)
	}
	return fmt.Errorf(
		"channel '%s' not found. Make sure you are a member of this channel",
		sanitized,
	)
}

// channelFromSlack converts a slack-go Channel to our internal Channel
// type.
func channelFromSlack(sc slackgo.Channel) Channel {
	return Channel{
		ID:              sc.ID,
		Name:            sc.Name,
		User:            sc.User,
		IsChannel:       sc.IsChannel,
		IsGroup:         sc.IsGroup,
		IsIM:            sc.IsIM,
		IsMPIM:          sc.IsMpIM,
		IsPrivate:       sc.IsPrivate,
		Created:         int64(sc.Created),
		IsArchived:      sc.IsArchived,
		IsGeneral:       sc.IsGeneral,
		Unlinked:        sc.Unlinked,
		NameNormalized:  sc.NameNormalized,
		IsShared:        sc.IsShared,
		IsExtShared:     sc.IsExtShared,
		IsOrgShared:     sc.IsOrgShared,
		IsMember:        sc.IsMember,
		NumMembers:      sc.NumMembers,
		UnreadCount:     sc.UnreadCount,
		UnreadCountDisp: sc.UnreadCountDisplay,
		LastRead:        sc.LastRead,
		Topic: &TopicPurpose{
			Value:   sc.Topic.Value,
			Creator: sc.Topic.Creator,
			LastSet: int64(sc.Topic.LastSet),
		},
		Purpose: &TopicPurpose{
			Value:   sc.Purpose.Value,
			Creator: sc.Purpose.Creator,
			LastSet: int64(sc.Purpose.LastSet),
		},
	}
}

// channelDetailFromSlack converts a slack-go Channel to our internal
// ChannelDetail type.
func channelDetailFromSlack(sc slackgo.Channel) ChannelDetail {
	return ChannelDetail{
		ID:         sc.ID,
		Name:       sc.Name,
		IsPrivate:  sc.IsPrivate,
		IsArchived: sc.IsArchived,
		Created:    int64(sc.Created),
		NumMembers: sc.NumMembers,
		Topic: &TopicPurpose{
			Value:   sc.Topic.Value,
			Creator: sc.Topic.Creator,
			LastSet: int64(sc.Topic.LastSet),
		},
		Purpose: &TopicPurpose{
			Value:   sc.Purpose.Value,
			Creator: sc.Purpose.Creator,
			LastSet: int64(sc.Purpose.LastSet),
		},
	}
}

// buildUnreadChannel merges a Channel with optional slack-go Channel
// info and the computed unread count.
func buildUnreadChannel(ch Channel, info *slackgo.Channel, unreadCount int) Channel {
	result := ch
	result.UnreadCount = unreadCount
	result.UnreadCountDisp = unreadCount

	if info != nil {
		if result.Name == "" {
			result.Name = info.Name
		}
		if result.LastRead == "" {
			result.LastRead = info.LastRead
		}
	}

	return result
}
