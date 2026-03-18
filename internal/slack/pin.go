package slack

import (
	"fmt"

	slackgo "github.com/slack-go/slack"
)

// PinOps groups pin-related API calls.
type PinOps struct {
	api        *slackgo.Client
	channelOps *ChannelOps
}

func newPinOps(api *slackgo.Client, channelOps *ChannelOps) *PinOps {
	return &PinOps{
		api:        api,
		channelOps: channelOps,
	}
}

// AddPin pins a message in the specified channel. The channel parameter
// accepts a name or ID.
func (p *PinOps) AddPin(channel, timestamp string) error {
	channelID, err := p.channelOps.ResolveChannelID(channel)
	if err != nil {
		return err
	}

	ref := slackgo.NewRefToMessage(channelID, timestamp)
	if err := p.api.AddPin(channelID, ref); err != nil {
		return fmt.Errorf("add pin: %w", err)
	}
	return nil
}

// RemovePin removes a pin from a message in the specified channel. The
// channel parameter accepts a name or ID.
func (p *PinOps) RemovePin(channel, timestamp string) error {
	channelID, err := p.channelOps.ResolveChannelID(channel)
	if err != nil {
		return err
	}

	ref := slackgo.NewRefToMessage(channelID, timestamp)
	if err := p.api.RemovePin(channelID, ref); err != nil {
		return fmt.Errorf("remove pin: %w", err)
	}
	return nil
}

// ListPins returns all pinned items in the specified channel. The
// channel parameter accepts a name or ID.
func (p *PinOps) ListPins(channel string) ([]PinnedItem, error) {
	channelID, err := p.channelOps.ResolveChannelID(channel)
	if err != nil {
		return nil, err
	}

	items, _, err := p.api.ListPins(channelID)
	if err != nil {
		return nil, fmt.Errorf("list pins: %w", err)
	}

	pinned := make([]PinnedItem, 0, len(items))
	for _, item := range items {
		pi := PinnedItem{
			Type: item.Type,
		}
		if item.Message != nil {
			pi.Message = &PinnedMessage{
				Text: item.Message.Text,
				User: item.Message.User,
				TS:   item.Message.Timestamp,
			}
		}
		pinned = append(pinned, pi)
	}
	return pinned, nil
}
