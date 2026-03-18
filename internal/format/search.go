package format

import (
	"fmt"

	"github.com/fatih/color"
)

// SearchMatchInfo holds the fields needed to display a single search match.
type SearchMatchInfo struct {
	Channel   string
	ChannelID string
	Username  string
	Timestamp string
	Text      string
	Permalink string
}

// SearchFormatOpts holds options for formatting search results.
type SearchFormatOpts struct {
	Query      string
	Matches    []SearchMatchInfo
	TotalCount int
	Page       int
	PageCount  int
}

// FormatSearchResults prints search results in the requested format.
func FormatSearchResults(opts SearchFormatOpts, f Format) {
	switch f {
	case JSON:
		type jsonMatch struct {
			Channel   string `json:"channel"`
			Username  string `json:"username"`
			Timestamp string `json:"timestamp"`
			Text      string `json:"text"`
			Permalink string `json:"permalink"`
		}
		type jsonOut struct {
			Query      string      `json:"query"`
			TotalCount int         `json:"total_count"`
			Page       int         `json:"page"`
			PageCount  int         `json:"page_count"`
			Matches    []jsonMatch `json:"matches"`
		}
		matches := make([]jsonMatch, len(opts.Matches))
		for i, m := range opts.Matches {
			matches[i] = jsonMatch{
				Channel:   channelOrDefault(m.Channel, m.ChannelID),
				Username:  usernameOrDefault(m.Username),
				Timestamp: m.Timestamp,
				Text:      textOrDefault(m.Text),
				Permalink: m.Permalink,
			}
		}
		PrintJSON(jsonOut{
			Query:      opts.Query,
			TotalCount: opts.TotalCount,
			Page:       opts.Page,
			PageCount:  opts.PageCount,
			Matches:    matches,
		})
	case Simple:
		if len(opts.Matches) == 0 {
			fmt.Println("No messages found")
			return
		}
		for _, m := range opts.Matches {
			ch := channelOrDefault(m.Channel, m.ChannelID)
			fmt.Printf("[%s] %s (%s): %s\n",
				ch, usernameOrDefault(m.Username), m.Timestamp, textOrDefault(m.Text))
		}
		if opts.TotalCount > len(opts.Matches) {
			fmt.Printf("... and %d more match(es)\n", opts.TotalCount-len(opts.Matches))
		}
	default:
		bold := color.New(color.Bold)
		gray := color.New(color.FgHiBlack)
		cyan := color.New(color.FgCyan)
		blue := color.New(color.FgBlue)
		green := color.New(color.FgGreen)

		bold.Printf("\nSearch results for \"%s\" (%d matches)\n", opts.Query, opts.TotalCount)

		if len(opts.Matches) == 0 {
			color.Yellow("No messages found")
			return
		}

		if opts.PageCount > 1 {
			gray.Printf("Page %d/%d\n", opts.Page, opts.PageCount)
		}

		fmt.Println()
		for _, m := range opts.Matches {
			ch := channelOrDefault(m.Channel, m.ChannelID)
			fmt.Printf("%s %s %s\n",
				gray.Sprintf("[%s]", m.Timestamp),
				blue.Sprint(ch),
				cyan.Sprint(usernameOrDefault(m.Username)))
			fmt.Println(textOrDefault(m.Text))
			if m.Permalink != "" {
				gray.Println(m.Permalink)
			}
			fmt.Println()
		}

		green.Printf("Displayed %d of %d match(es)\n", len(opts.Matches), opts.TotalCount)
	}
}

func channelOrDefault(name, id string) string {
	if name != "" {
		return "#" + name
	}
	if id != "" {
		return id
	}
	return "unknown"
}

func usernameOrDefault(u string) string {
	if u == "" {
		return "Unknown"
	}
	return u
}
