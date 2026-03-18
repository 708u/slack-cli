package slack

import (
	"sort"
	"strconv"

	slackgo "github.com/slack-go/slack"
)

const (
	unreadSearchQuery    = "is:unread"
	unreadSearchPageSize = 100
)

func newSearchOps(api *slackgo.Client) *SearchOps {
	return &SearchOps{api: api}
}

// SearchMessages searches workspace messages with the given query and
// options. Zero-value fields in opts fall back to Slack API defaults
// (sort=score, sortDir=desc, count=20, page=1).
func (s *SearchOps) SearchMessages(query string, opts SearchMessagesOptions) (*SearchResult, error) {
	params := slackgo.SearchParameters{
		Sort:          opts.Sort,
		SortDirection: opts.SortDir,
		Count:         opts.Count,
		Page:          opts.Page,
	}
	if params.Sort == "" {
		params.Sort = "score"
	}
	if params.SortDirection == "" {
		params.SortDirection = "desc"
	}
	if params.Count == 0 {
		params.Count = 20
	}
	if params.Page == 0 {
		params.Page = 1
	}

	resp, err := s.api.SearchMessages(query, params)
	if err != nil {
		return nil, wrapSlackError("search messages", err)
	}

	matches := make([]SearchMatch, 0, len(resp.Matches))
	for _, m := range resp.Matches {
		matches = append(matches, SearchMatch{
			Text:      m.Text,
			User:      m.User,
			Username:  m.Username,
			TS:        m.Timestamp,
			Permalink: m.Permalink,
			Channel: SearchMatchChannel{
				ID:   m.Channel.ID,
				Name: m.Channel.Name,
			},
		})
	}

	return &SearchResult{
		Query:      query,
		Matches:    matches,
		TotalCount: resp.Pagination.TotalCount,
		Page:       resp.Pagination.Page,
		PageCount:  resp.Pagination.PageCount,
	}, nil
}

// ListUnreadChannels searches for "is:unread" messages across all pages
// and aggregates the results by channel, returning channels sorted by
// latest message timestamp (descending).
func (s *SearchOps) ListUnreadChannels() ([]Channel, error) {
	firstPage, err := s.fetchSearchPage(unreadSearchQuery, 1, unreadSearchPageSize)
	if err != nil {
		return nil, err
	}

	matches := make([]slackgo.SearchMessage, 0, len(firstPage.Matches))
	matches = append(matches, firstPage.Matches...)
	pageCount := firstPage.Pagination.PageCount

	for page := 2; page <= pageCount; page++ {
		resp, err := s.fetchSearchPage(unreadSearchQuery, page, unreadSearchPageSize)
		if err != nil {
			return nil, err
		}
		matches = append(matches, resp.Matches...)
	}

	return aggregateUnreadChannels(matches), nil
}

// fetchSearchPage retrieves a single page of search results sorted by
// timestamp descending.
func (s *SearchOps) fetchSearchPage(query string, page, count int) (*slackgo.SearchMessages, error) {
	params := slackgo.SearchParameters{
		Sort:          "timestamp",
		SortDirection: "desc",
		Count:         count,
		Page:          page,
	}

	resp, err := s.api.SearchMessages(query, params)
	if err != nil {
		return nil, wrapSlackError("search messages", err)
	}
	return resp, nil
}

// aggregateUnreadChannels groups search matches by channel, counting
// unread messages per channel and tracking the latest message timestamp.
func aggregateUnreadChannels(matches []slackgo.SearchMessage) []Channel {
	type entry struct {
		channel     Channel
		unreadCount int
	}

	channels := make(map[string]*entry)

	for _, m := range matches {
		id := m.Channel.ID
		if id == "" {
			continue
		}

		e, ok := channels[id]
		if ok {
			e.unreadCount++
			e.channel.UnreadCount = e.unreadCount
			e.channel.UnreadCountDisp = e.unreadCount
			e.channel.LastRead = maxSlackTimestamp(e.channel.LastRead, m.Timestamp)
			continue
		}

		channels[id] = &entry{
			channel: Channel{
				ID:              id,
				Name:            m.Channel.Name,
				IsChannel:       !m.Channel.IsPrivate && !m.Channel.IsMPIM,
				IsMPIM:          m.Channel.IsMPIM,
				IsPrivate:       m.Channel.IsPrivate,
				UnreadCount:     1,
				UnreadCountDisp: 1,
				LastRead:        m.Timestamp,
			},
			unreadCount: 1,
		}
	}

	result := make([]Channel, 0, len(channels))
	for _, e := range channels {
		result = append(result, e.channel)
	}

	sort.Slice(result, func(i, j int) bool {
		ti, _ := strconv.ParseFloat(result[i].LastRead, 64)
		tj, _ := strconv.ParseFloat(result[j].LastRead, 64)
		return ti > tj
	})

	return result
}

// maxSlackTimestamp returns whichever of current and candidate is the
// larger Slack timestamp string (dot-separated epoch). If either is
// empty, the other is returned.
func maxSlackTimestamp(current, candidate string) string {
	if candidate == "" {
		return current
	}
	if current == "" {
		return candidate
	}

	cv, _ := strconv.ParseFloat(candidate, 64)
	cu, _ := strconv.ParseFloat(current, 64)
	if cv > cu {
		return candidate
	}
	return current
}
