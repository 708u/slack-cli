package slack

import (
	slackgo "github.com/slack-go/slack"
)

// Client wraps the slack-go API client and delegates to domain-specific
// operation structs. It supports separate bot and user tokens; APIs
// that require a user token (search, stars, reminders) use the user
// client automatically.
type Client struct {
	botAPI  *slackgo.Client
	userAPI *slackgo.Client

	channelOps  *ChannelOps
	messageOps  *MessageOps
	userOps     *UserOps
	searchOps   *SearchOps
	reactionOps *ReactionOps
	pinOps      *PinOps
	reminderOps *ReminderOps
	starOps     *StarOps
	fileOps     *FileOps
	canvasOps   *CanvasOps
}

// NewClient creates a Client. botToken and userToken are both optional
// but at least one must be non-empty.
//   - If only botToken: user-token-only commands will fail with a clear error.
//   - If only userToken: used for all operations.
//   - If both: bot token for general ops, user token for search/stars/reminders.
func NewClient(botToken, userToken string) *Client {
	// primary is used for most operations.
	primary := botToken
	if primary == "" {
		primary = userToken
	}

	// userOnly is used for APIs that require a user token.
	userOnly := userToken
	if userOnly == "" {
		userOnly = primary
	}

	api := slackgo.New(primary)
	userAPI := api
	if userOnly != primary {
		userAPI = slackgo.New(userOnly)
	}

	// channelAPI uses the user token when available for broader
	// channel visibility (e.g. private channels with groups:read).
	channelAPI := userAPI

	c := &Client{botAPI: api, userAPI: userAPI}

	c.channelOps = newChannelOps(channelAPI)
	c.messageOps = newMessageOps(api, userAPI, c.channelOps)
	c.userOps = newUserOps(api)
	c.searchOps = newSearchOps(userAPI)
	c.reactionOps = newReactionOps(api, c.channelOps)
	c.pinOps = newPinOps(api, c.channelOps)
	c.reminderOps = newReminderOps(userAPI, userOnly)
	c.starOps = newStarOps(userAPI)
	c.fileOps = newFileOps(api, c.channelOps)
	c.canvasOps = newCanvasOps(api, primary, c.channelOps)

	return c
}

// NOTE: ChannelOps is defined in channel.go.

// MessageOps groups message read/write API calls. It holds a
// fallbackAPI (user token) to retry writes when the primary (bot)
// token gets channel_not_found on private channels.
type MessageOps struct {
	api         *slackgo.Client
	fallbackAPI *slackgo.Client
	channelOps  *ChannelOps
}

// UserOps groups user-related API calls.
type UserOps struct {
	api *slackgo.Client
}

// SearchOps groups search-related API calls.
type SearchOps struct {
	api *slackgo.Client
}

// NOTE: ReactionOps is defined in reaction.go.
// NOTE: PinOps is defined in pin.go.
// NOTE: ReminderOps is defined in reminder.go.
// NOTE: StarOps is defined in star.go.
// NOTE: FileOps is defined in file.go.
// NOTE: CanvasOps is defined in canvas.go.

// ---------------------------------------------------------------------------
// Channel operations
// ---------------------------------------------------------------------------

func (c *Client) ListChannels(opts ListChannelsOptions) ([]Channel, error) {
	return c.channelOps.ListChannels(opts)
}

func (c *Client) GetChannelDetail(channelNameOrID string) (*ChannelDetail, error) {
	return c.channelOps.GetChannelDetail(channelNameOrID)
}

func (c *Client) SetTopic(channelNameOrID, topic string) error {
	return c.channelOps.SetTopic(channelNameOrID, topic)
}

func (c *Client) SetPurpose(channelNameOrID, purpose string) error {
	return c.channelOps.SetPurpose(channelNameOrID, purpose)
}

func (c *Client) JoinChannel(channelNameOrID string) error {
	return c.channelOps.JoinChannel(channelNameOrID)
}

func (c *Client) LeaveChannel(channelNameOrID string) error {
	return c.channelOps.LeaveChannel(channelNameOrID)
}

func (c *Client) InviteToChannel(channelNameOrID string, userIDs []string, force bool) error {
	return c.channelOps.InviteToChannel(channelNameOrID, userIDs, force)
}

func (c *Client) GetChannelMembers(channelNameOrID string, opts *ChannelMembersOptions) (*ChannelMembersResult, error) {
	if opts == nil {
		opts = &ChannelMembersOptions{}
	}
	return c.channelOps.GetChannelMembers(channelNameOrID, *opts)
}

func (c *Client) ListUnreadChannels() ([]Channel, error) {
	return c.channelOps.ListUnreadChannels()
}

func (c *Client) FetchUserChannels() ([]Channel, error) {
	return c.channelOps.FetchUserChannels()
}

func (c *Client) ResolveChannelID(nameOrID string) (string, error) {
	return c.channelOps.ResolveChannelID(nameOrID)
}

// ---------------------------------------------------------------------------
// Message operations
// ---------------------------------------------------------------------------

func (c *Client) SendMessage(channel, text, threadTS string) error {
	return c.messageOps.SendMessage(channel, text, threadTS)
}

func (c *Client) SendEphemeralMessage(channel, user, text, threadTS string) error {
	return c.messageOps.SendEphemeralMessage(channel, user, text, threadTS)
}

func (c *Client) ScheduleMessage(channel, text string, postAt int64, threadTS string) error {
	return c.messageOps.ScheduleMessage(channel, text, postAt, threadTS)
}

func (c *Client) UpdateMessage(channel, ts, text string) error {
	return c.messageOps.UpdateMessage(channel, ts, text)
}

func (c *Client) DeleteMessage(channel, ts string) error {
	return c.messageOps.DeleteMessage(channel, ts)
}

func (c *Client) ListScheduledMessages(channel string, limit int) ([]ScheduledMessage, error) {
	return c.messageOps.ListScheduledMessages(channel, limit)
}

func (c *Client) CancelScheduledMessage(channel, scheduledMessageID string) error {
	return c.messageOps.CancelScheduledMessage(channel, scheduledMessageID)
}

func (c *Client) GetHistory(channel string, opts HistoryOptions) (*HistoryResult, error) {
	return c.messageOps.GetHistory(channel, opts)
}

func (c *Client) GetThreadHistory(channel, threadTS string) (*HistoryResult, error) {
	return c.messageOps.GetThreadHistory(channel, threadTS)
}

func (c *Client) GetChannelUnread(channelNameOrID string) (*ChannelUnreadResult, error) {
	return c.messageOps.GetChannelUnread(channelNameOrID)
}

func (c *Client) MarkAsRead(channelID string) error {
	return c.messageOps.MarkAsRead(channelID)
}

func (c *Client) GetPermalink(channel, messageTS string) (string, error) {
	return c.messageOps.GetPermalink(channel, messageTS)
}

func (c *Client) GetPermalinks(channel string, messageTimestamps []string) (map[string]string, error) {
	return c.messageOps.GetPermalinks(channel, messageTimestamps)
}

// ---------------------------------------------------------------------------
// File operations
// ---------------------------------------------------------------------------

func (c *Client) UploadFile(opts UploadFileOptions) error {
	return c.fileOps.UploadFile(opts)
}

// ---------------------------------------------------------------------------
// Reaction operations
// ---------------------------------------------------------------------------

func (c *Client) AddReaction(channel, timestamp, emoji string) error {
	return c.reactionOps.AddReaction(channel, timestamp, emoji)
}

func (c *Client) RemoveReaction(channel, timestamp, emoji string) error {
	return c.reactionOps.RemoveReaction(channel, timestamp, emoji)
}

// ---------------------------------------------------------------------------
// Pin operations
// ---------------------------------------------------------------------------

func (c *Client) AddPin(channel, timestamp string) error {
	return c.pinOps.AddPin(channel, timestamp)
}

func (c *Client) RemovePin(channel, timestamp string) error {
	return c.pinOps.RemovePin(channel, timestamp)
}

func (c *Client) ListPins(channel string) ([]PinnedItem, error) {
	return c.pinOps.ListPins(channel)
}

// ---------------------------------------------------------------------------
// User operations
// ---------------------------------------------------------------------------

func (c *Client) ListUsers(limit int) ([]SlackUser, error) {
	return c.userOps.ListUsers(limit)
}

func (c *Client) GetUserInfo(userID string) (*SlackUser, error) {
	return c.userOps.GetUserInfo(userID)
}

func (c *Client) LookupUserByEmail(email string) (*SlackUser, error) {
	return c.userOps.LookupByEmail(email)
}

func (c *Client) OpenDMChannel(userID string) (string, error) {
	return c.userOps.OpenDMChannel(userID)
}

func (c *Client) GetUserPresence(userID string) (*UserPresence, error) {
	return c.userOps.GetPresence(userID)
}

func (c *Client) ResolveUserIDByName(username string) (string, error) {
	return c.userOps.ResolveUserIDByName(username)
}

// ---------------------------------------------------------------------------
// Search operations
// ---------------------------------------------------------------------------

func (c *Client) SearchMessages(query string, opts *SearchMessagesOptions) (*SearchResult, error) {
	o := SearchMessagesOptions{}
	if opts != nil {
		o = *opts
	}
	return c.searchOps.SearchMessages(query, o)
}

// ---------------------------------------------------------------------------
// Reminder operations
// ---------------------------------------------------------------------------

func (c *Client) AddReminder(text string, time int64) (*Reminder, error) {
	return c.reminderOps.AddReminder(text, time)
}

func (c *Client) ListReminders() ([]Reminder, error) {
	return c.reminderOps.ListReminders()
}

func (c *Client) DeleteReminder(reminderID string) error {
	return c.reminderOps.DeleteReminder(reminderID)
}

func (c *Client) CompleteReminder(reminderID string) error {
	return c.reminderOps.CompleteReminder(reminderID)
}

// ---------------------------------------------------------------------------
// Star operations
// ---------------------------------------------------------------------------

func (c *Client) AddStar(channel, timestamp string) error {
	return c.starOps.AddStar(channel, timestamp)
}

func (c *Client) ListStars(count int) (*StarListResult, error) {
	return c.starOps.ListStars(count)
}

func (c *Client) RemoveStar(channel, timestamp string) error {
	return c.starOps.RemoveStar(channel, timestamp)
}

// ---------------------------------------------------------------------------
// Canvas operations
// ---------------------------------------------------------------------------

func (c *Client) ReadCanvas(canvasID string) ([]CanvasSection, error) {
	return c.canvasOps.ReadCanvas(canvasID)
}

func (c *Client) ListCanvases(channel string) ([]CanvasFile, error) {
	return c.canvasOps.ListCanvases(channel)
}
