package slack

import (
	"fmt"
	"strings"

	slackgo "github.com/slack-go/slack"
)

// ReactionOps groups reaction-related API calls.
type ReactionOps struct {
	api        *slackgo.Client
	channelOps *ChannelOps
}

func newReactionOps(api *slackgo.Client, channelOps *ChannelOps) *ReactionOps {
	return &ReactionOps{
		api:        api,
		channelOps: channelOps,
	}
}

// stripColons removes leading and trailing colons from an emoji name
// (e.g. ":thumbsup:" -> "thumbsup").
func stripColons(emoji string) string {
	return strings.TrimSuffix(strings.TrimPrefix(emoji, ":"), ":")
}

// AddReaction adds an emoji reaction to the specified message. The
// emoji name is normalised by stripping surrounding colons. The channel
// parameter accepts a name or ID.
func (r *ReactionOps) AddReaction(channel, timestamp, emoji string) error {
	channelID, err := r.channelOps.ResolveChannelID(channel)
	if err != nil {
		return err
	}

	ref := slackgo.NewRefToMessage(channelID, timestamp)
	if err := r.api.AddReaction(stripColons(emoji), ref); err != nil {
		return fmt.Errorf("add reaction: %w", err)
	}
	return nil
}

// RemoveReaction removes an emoji reaction from the specified message.
// The channel parameter accepts a name or ID.
func (r *ReactionOps) RemoveReaction(channel, timestamp, emoji string) error {
	channelID, err := r.channelOps.ResolveChannelID(channel)
	if err != nil {
		return err
	}

	ref := slackgo.NewRefToMessage(channelID, timestamp)
	if err := r.api.RemoveReaction(stripColons(emoji), ref); err != nil {
		return fmt.Errorf("remove reaction: %w", err)
	}
	return nil
}
