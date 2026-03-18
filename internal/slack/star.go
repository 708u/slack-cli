package slack

import (
	slackgo "github.com/slack-go/slack"
)

// StarOps groups star/bookmark-related API calls.
type StarOps struct {
	api *slackgo.Client
}

func newStarOps(api *slackgo.Client) *StarOps {
	return &StarOps{api: api}
}

// AddStar stars (bookmarks) a message identified by channel and
// timestamp.
func (s *StarOps) AddStar(channel, timestamp string) error {
	ref := slackgo.NewRefToMessage(channel, timestamp)
	if err := s.api.AddStar(channel, ref); err != nil {
		return wrapSlackError("add star", err)
	}
	return nil
}

// ListStars returns starred items for the authenticated user. count
// controls how many items to return per page.
func (s *StarOps) ListStars(count int) (*StarListResult, error) {
	if count <= 0 {
		count = 100
	}

	params := slackgo.StarsParameters{
		Count: count,
		Page:  1,
	}

	items, _, err := s.api.ListStars(params)
	if err != nil {
		return nil, wrapSlackError("list stars", err)
	}

	starred := make([]StarredItem, 0, len(items))
	for _, item := range items {
		si := StarredItem{
			Type:    item.Type,
			Channel: item.Channel,
		}
		if item.Message != nil {
			si.Message = StarMessage{
				Text: item.Message.Text,
				TS:   item.Message.Timestamp,
			}
		}
		starred = append(starred, si)
	}

	return &StarListResult{Items: starred}, nil
}

// RemoveStar removes a star from a message identified by channel and
// timestamp.
func (s *StarOps) RemoveStar(channel, timestamp string) error {
	ref := slackgo.NewRefToMessage(channel, timestamp)
	if err := s.api.RemoveStar(channel, ref); err != nil {
		return wrapSlackError("remove star", err)
	}
	return nil
}
