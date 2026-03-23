package cmd

import (
	"fmt"

	"github.com/708u/slack-cli/internal/format"
	"github.com/708u/slack-cli/internal/slack"
	"github.com/708u/slack-cli/internal/util"
)

// SearchCmd searches messages in a Slack workspace.
type SearchCmd struct {
	Query   string `name:"query" short:"q" help:"Search query" required:""`
	Sort    string `name:"sort" help:"Sort: score or timestamp" default:"score"`
	SortDir string `name:"sort-dir" help:"Sort direction: asc or desc" default:"desc"`
	Number  int    `name:"number" short:"n" help:"Results per page" default:"20"`
	Page    int    `name:"page" help:"Page number" default:"1"`
	Format  string `name:"format" enum:"table,simple,json" default:"table"`
}

// Run executes the search command.
func (c *SearchCmd) Run(client *slack.Client) error {
	result, err := client.SearchMessages(c.Query, &slack.SearchMessagesOptions{
		Sort:    c.Sort,
		SortDir: c.SortDir,
		Count:   c.Number,
		Page:    c.Page,
	})
	if err != nil {
		return err
	}

	if len(result.Matches) == 0 {
		fmt.Printf("No messages found for query %q\n", c.Query)
		return nil
	}

	infos := make([]format.SearchMatchInfo, len(result.Matches))
	for i, m := range result.Matches {
		infos[i] = format.SearchMatchInfo{
			Channel:   util.SanitizeTerminalText(m.Channel.Name),
			ChannelID: m.Channel.ID,
			Username:  util.SanitizeTerminalText(m.Username),
			Timestamp: util.FormatSlackTimestamp(m.TS),
			Text:      util.SanitizeTerminalText(m.Text),
			Permalink: m.Permalink,
		}
	}

	f := format.ParseFormat(c.Format)
	format.FormatSearchResults(format.SearchFormatOpts{
		Query:      result.Query,
		Matches:    infos,
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageCount:  result.PageCount,
	}, f)

	return nil
}
