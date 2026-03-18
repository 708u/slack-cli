package slack

import (
	"fmt"
	"strconv"
	"time"

	"github.com/708u/slack-cli/internal/util"
	slackgo "github.com/slack-go/slack"
)

func newMessageOps(api *slackgo.Client, channelOps *ChannelOps) *MessageOps {
	return &MessageOps{
		api:        api,
		channelOps: channelOps,
	}
}

// SendMessage posts a message to the given channel. If threadTS is
// non-empty the message is posted as a thread reply.
func (m *MessageOps) SendMessage(channel, text string, threadTS string) error {
	opts := []slackgo.MsgOption{
		slackgo.MsgOptionText(text, false),
	}
	if threadTS != "" {
		opts = append(opts, slackgo.MsgOptionTS(threadTS))
	}

	_, _, err := m.api.PostMessage(channel, opts...)
	return err
}

// SendEphemeralMessage posts an ephemeral message visible only to the
// specified user. If threadTS is non-empty the message appears in that
// thread.
func (m *MessageOps) SendEphemeralMessage(channel, user, text string, threadTS string) error {
	opts := []slackgo.MsgOption{
		slackgo.MsgOptionText(text, false),
	}
	if threadTS != "" {
		opts = append(opts, slackgo.MsgOptionTS(threadTS))
	}

	_, err := m.api.PostEphemeral(channel, user, opts...)
	return err
}

// ScheduleMessage schedules a message to be sent at postAt (Unix epoch
// seconds). If threadTS is non-empty the scheduled message will be a
// thread reply.
func (m *MessageOps) ScheduleMessage(channel, text string, postAt int64, threadTS string) error {
	postAtStr := strconv.FormatInt(postAt, 10)
	opts := []slackgo.MsgOption{
		slackgo.MsgOptionText(text, false),
		slackgo.MsgOptionSchedule(postAtStr),
	}
	if threadTS != "" {
		opts = append(opts, slackgo.MsgOptionTS(threadTS))
	}

	_, _, err := m.api.ScheduleMessage(channel, postAtStr, opts...)
	return err
}

// UpdateMessage edits an existing message identified by channel and ts.
func (m *MessageOps) UpdateMessage(channel, ts, text string) error {
	channelID, err := m.channelOps.ResolveChannelID(channel)
	if err != nil {
		return err
	}

	_, _, _, err = m.api.UpdateMessage(
		channelID,
		ts,
		slackgo.MsgOptionText(text, false),
	)
	return err
}

// DeleteMessage removes a message identified by channel and ts.
func (m *MessageOps) DeleteMessage(channel, ts string) error {
	channelID, err := m.channelOps.ResolveChannelID(channel)
	if err != nil {
		return err
	}

	_, _, err = m.api.DeleteMessage(channelID, ts)
	return err
}

// ListScheduledMessages returns scheduled messages for the given
// channel. Pass an empty channel string to list across all channels.
func (m *MessageOps) ListScheduledMessages(channel string, limit int) ([]ScheduledMessage, error) {
	var channelID string
	if channel != "" {
		var err error
		channelID, err = m.channelOps.ResolveChannelID(channel)
		if err != nil {
			return nil, err
		}
	}

	msgs, _, err := m.api.GetScheduledMessages(&slackgo.GetScheduledMessagesParameters{
		Channel: channelID,
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}

	result := make([]ScheduledMessage, 0, len(msgs))
	for _, sm := range msgs {
		result = append(result, scheduledMessageFromSlack(sm))
	}
	return result, nil
}

// CancelScheduledMessage deletes a pending scheduled message.
func (m *MessageOps) CancelScheduledMessage(channel, scheduledMsgID string) error {
	channelID, err := m.channelOps.ResolveChannelID(channel)
	if err != nil {
		return err
	}

	_, err = m.api.DeleteScheduledMessage(&slackgo.DeleteScheduledMessageParameters{
		Channel:            channelID,
		ScheduledMessageID: scheduledMsgID,
	})
	return err
}

// GetHistory retrieves conversation history for the given channel.
// User IDs found in messages are resolved to display names.
func (m *MessageOps) GetHistory(channel string, opts HistoryOptions) (*HistoryResult, error) {
	channelID, err := m.channelOps.ResolveChannelID(channel)
	if err != nil {
		return nil, err
	}

	resp, err := m.api.GetConversationHistory(&slackgo.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     opts.Limit,
		Oldest:    opts.Oldest,
	})
	if err != nil {
		return nil, err
	}

	messages := messagesFromSlack(resp.Messages)
	userIDs := extractAllUserIDs(messages)
	users, err := m.fetchUserNames(userIDs)
	if err != nil {
		return nil, err
	}

	return &HistoryResult{
		Messages: messages,
		Users:    users,
	}, nil
}

// GetThreadHistory retrieves all replies in a thread, paginating
// through the full reply set.
func (m *MessageOps) GetThreadHistory(channel, threadTS string) (*HistoryResult, error) {
	channelID, err := m.channelOps.ResolveChannelID(channel)
	if err != nil {
		return nil, err
	}

	var messages []Message
	var cursor string

	for {
		msgs, _, nextCursor, err := m.api.GetConversationReplies(&slackgo.GetConversationRepliesParameters{
			ChannelID: channelID,
			Timestamp: threadTS,
			Cursor:    cursor,
		})
		if err != nil {
			return nil, err
		}

		messages = append(messages, messagesFromSlack(msgs)...)

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	userIDs := extractAllUserIDs(messages)
	users, err := m.fetchUserNames(userIDs)
	if err != nil {
		return nil, err
	}

	return &HistoryResult{
		Messages: messages,
		Users:    users,
	}, nil
}

// unreadPreviewLimit is the maximum number of unread messages to
// include in a ChannelUnreadResult preview.
const unreadPreviewLimit = 10

// GetChannelUnread returns unread message information for the given
// channel, including a preview of up to unreadPreviewLimit messages.
func (m *MessageOps) GetChannelUnread(nameOrID string) (*ChannelUnreadResult, error) {
	channelID, err := m.channelOps.ResolveChannelID(nameOrID)
	if err != nil {
		return nil, err
	}

	ch, err := m.api.GetConversationInfo(&slackgo.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return nil, err
	}

	summary, err := m.getUnreadSummary(channelID, ch.LastRead, unreadPreviewLimit)
	if err != nil {
		return nil, err
	}

	userIDs := extractAllUserIDs(summary.messages)
	users, err := m.fetchUserNames(userIDs)
	if err != nil {
		return nil, err
	}

	channel := channelFromSlack(*ch)
	channel.UnreadCount = summary.totalCount
	channel.UnreadCountDisp = summary.totalCount

	return &ChannelUnreadResult{
		Channel:               channel,
		Messages:              summary.messages,
		Users:                 users,
		TotalUnreadCount:      summary.totalCount,
		DisplayedMessageCount: len(summary.messages),
	}, nil
}

// MarkAsRead marks the conversation as read up to the current time.
func (m *MessageOps) MarkAsRead(channelID string) error {
	now := fmt.Sprintf("%.6f", float64(time.Now().UnixMilli())/1000.0)
	return m.api.MarkConversation(channelID, now)
}

// GetPermalink returns the permalink URL for a single message. An
// empty string is returned when the permalink cannot be obtained.
func (m *MessageOps) GetPermalink(channel, messageTS string) (string, error) {
	channelID, err := m.channelOps.ResolveChannelID(channel)
	if err != nil {
		return "", err
	}

	link, err := m.api.GetPermalink(&slackgo.PermalinkParameters{
		Channel: channelID,
		Ts:      messageTS,
	})
	if err != nil {
		return "", err
	}
	return link, nil
}

// GetPermalinks returns permalink URLs for multiple messages keyed by
// timestamp. Timestamps that fail to resolve are silently omitted.
func (m *MessageOps) GetPermalinks(channel string, timestamps []string) (map[string]string, error) {
	result := make(map[string]string, len(timestamps))
	if len(timestamps) == 0 {
		return result, nil
	}

	channelID, err := m.channelOps.ResolveChannelID(channel)
	if err != nil {
		return nil, err
	}

	for _, ts := range timestamps {
		link, err := m.api.GetPermalink(&slackgo.PermalinkParameters{
			Channel: channelID,
			Ts:      ts,
		})
		if err != nil {
			continue
		}
		result[ts] = link
	}
	return result, nil
}

// fetchUserNames resolves a set of user IDs to display names by
// calling the Slack users.info API for each ID. On failure the raw
// user ID is used as the display name.
func (m *MessageOps) fetchUserNames(userIDs []string) (map[string]string, error) {
	users := make(map[string]string, len(userIDs))
	for _, id := range userIDs {
		info, err := m.api.GetUserInfo(id)
		if err != nil {
			users[id] = id
			continue
		}
		if info.Name != "" {
			users[id] = info.Name
		} else {
			users[id] = id
		}
	}
	return users, nil
}

// unreadSummary holds the result of scanning unread messages.
type unreadSummary struct {
	totalCount int
	messages   []Message
}

// getUnreadSummary pages through conversation history from lastRead
// onward, counting all unread messages but only keeping up to
// previewLimit for display.
func (m *MessageOps) getUnreadSummary(channelID, lastRead string, previewLimit int) (*unreadSummary, error) {
	var (
		messages   []Message
		totalCount int
		cursor     string
	)

	for {
		resp, err := m.api.GetConversationHistory(&slackgo.GetConversationHistoryParameters{
			ChannelID: channelID,
			Oldest:    lastRead,
			Limit:     200,
			Cursor:    cursor,
		})
		if err != nil {
			return nil, err
		}

		page := messagesFromSlack(resp.Messages)
		totalCount += len(page)

		if remaining := previewLimit - len(messages); remaining > 0 {
			if len(page) <= remaining {
				messages = append(messages, page...)
			} else {
				messages = append(messages, page[:remaining]...)
			}
		}

		cursor = resp.ResponseMetaData.NextCursor
		if cursor == "" {
			break
		}
	}

	return &unreadSummary{
		totalCount: totalCount,
		messages:   messages,
	}, nil
}

// ---------------------------------------------------------------------------
// Type conversion helpers
// ---------------------------------------------------------------------------

// GetUser returns the message author's user ID. It satisfies the
// util.MessageLike interface.
func (msg *Message) GetUser() string { return msg.User }

// GetText returns the message body. It satisfies the
// util.MessageLike interface.
func (msg *Message) GetText() string { return msg.Text }

// messageFromSlack converts a slack-go Message into the internal
// Message type.
func messageFromSlack(sm slackgo.Message) Message {
	return Message{
		Type:       sm.Type,
		Text:       sm.Text,
		User:       sm.User,
		BotID:      sm.BotID,
		TS:         sm.Timestamp,
		ThreadTS:   sm.ThreadTimestamp,
		ReplyCount: sm.ReplyCount,
	}
}

// messagesFromSlack converts a slice of slack-go Messages.
func messagesFromSlack(sms []slackgo.Message) []Message {
	msgs := make([]Message, 0, len(sms))
	for _, sm := range sms {
		msgs = append(msgs, messageFromSlack(sm))
	}
	return msgs
}

// scheduledMessageFromSlack converts a slack-go ScheduledMessage into
// the internal type.
func scheduledMessageFromSlack(sm slackgo.ScheduledMessage) ScheduledMessage {
	return ScheduledMessage{
		ID:          sm.ID,
		ChannelID:   sm.Channel,
		PostAt:      int64(sm.PostAt),
		DateCreated: int64(sm.DateCreated),
		Text:        sm.Text,
	}
}

// extractAllUserIDs collects unique user IDs from message authors and
// <@USERID> mentions across all messages.
func extractAllUserIDs(messages []Message) []string {
	ifaces := make([]util.MessageLike, 0, len(messages))
	for i := range messages {
		ifaces = append(ifaces, &messages[i])
	}
	return util.ExtractAllUserIDs(ifaces)
}
