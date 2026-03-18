package format

import (
	"fmt"

	"github.com/fatih/color"
)

// ChannelInfo holds the fields needed to display a channel list row.
type ChannelInfo struct {
	ID      string
	Name    string
	Type    string
	Members int
	Created string
	Purpose string
}

// FormatChannelsList prints a list of channels in the requested format.
func FormatChannelsList(channels []ChannelInfo, f Format) {
	switch f {
	case JSON:
		type jsonChannel struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Type    string `json:"type"`
			Members int    `json:"members"`
			Created string `json:"created"`
			Purpose string `json:"purpose"`
		}
		out := make([]jsonChannel, len(channels))
		for i, ch := range channels {
			out[i] = jsonChannel{
				ID:      ch.ID,
				Name:    ch.Name,
				Type:    ch.Type,
				Members: ch.Members,
				Created: ch.Created,
				Purpose: ch.Purpose,
			}
		}
		PrintJSON(out)
	case Simple:
		for _, ch := range channels {
			fmt.Println(ch.Name)
		}
	default:
		bold := color.New(color.Bold)
		bold.Printf("%-18s%-10s%-9s%-13s%s\n", "Name", "Type", "Members", "Created", "Description")
		fmt.Println(Separator(65))
		for _, ch := range channels {
			purpose := truncate(ch.Purpose, 30)
			fmt.Printf("%-18s%-10s%-9d%-13s%s\n", ch.Name, ch.Type, ch.Members, ch.Created, purpose)
		}
	}
}

// ChannelDetailInfo holds the fields needed to display detailed channel info.
type ChannelDetailInfo struct {
	ID         string
	Name       string
	IsPrivate  bool
	IsArchived bool
	Created    string
	NumMembers int
	Topic      string
	Purpose    string
}

// FormatChannelInfo prints detailed information about a single channel.
func FormatChannelInfo(ch ChannelDetailInfo, f Format) {
	switch f {
	case JSON:
		out := struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			IsPrivate  bool   `json:"is_private"`
			IsArchived bool   `json:"is_archived"`
			Created    string `json:"created"`
			NumMembers int    `json:"num_members"`
			Topic      string `json:"topic"`
			Purpose    string `json:"purpose"`
		}{
			ID:         ch.ID,
			Name:       ch.Name,
			IsPrivate:  ch.IsPrivate,
			IsArchived: ch.IsArchived,
			Created:    ch.Created,
			NumMembers: ch.NumMembers,
			Topic:      ch.Topic,
			Purpose:    ch.Purpose,
		}
		PrintJSON(out)
	case Simple:
		fmt.Printf("%s (%s)\n", ch.Name, ch.ID)
		fmt.Printf("Topic: %s\n", ch.Topic)
		fmt.Printf("Purpose: %s\n", ch.Purpose)
		if ch.NumMembers > 0 {
			fmt.Printf("Members: %d\n", ch.NumMembers)
		}
	default:
		bold := color.New(color.Bold)
		gray := color.New(color.FgHiBlack)

		bold.Printf("\nChannel Info: #%s\n", ch.Name)
		fmt.Println()
		fmt.Printf("  %s %s\n", gray.Sprint("ID:"), ch.ID)
		fmt.Printf("  %s %s\n", gray.Sprint("Name:"), ch.Name)
		fmt.Printf("  %s %s\n", gray.Sprint("Private:"), boolYesNo(ch.IsPrivate))
		fmt.Printf("  %s %s\n", gray.Sprint("Archived:"), boolYesNo(ch.IsArchived))
		if ch.NumMembers > 0 {
			fmt.Printf("  %s %d\n", gray.Sprint("Members:"), ch.NumMembers)
		}
		fmt.Printf("  %s %s\n", gray.Sprint("Created:"), ch.Created)
		fmt.Println()
		fmt.Printf("  %s %s\n", gray.Sprint("Topic:"), ch.Topic)
		fmt.Printf("  %s %s\n", gray.Sprint("Purpose:"), ch.Purpose)
		fmt.Println()
	}
}

// UnreadChannelInfo holds the fields needed to display an unread-channels row.
type UnreadChannelInfo struct {
	ID          string
	Name        string
	UnreadCount int
	LastRead    string
}

// FormatUnreadChannels prints a list of channels with unread counts.
// When countOnly is true, the output shows per-channel counts and a total.
func FormatUnreadChannels(channels []UnreadChannelInfo, f Format, countOnly bool) {
	switch f {
	case JSON:
		type jsonUnread struct {
			Channel     string `json:"channel"`
			ChannelID   string `json:"channel_id"`
			UnreadCount int    `json:"unread_count"`
		}
		out := make([]jsonUnread, len(channels))
		for i, ch := range channels {
			out[i] = jsonUnread{
				Channel:     ch.Name,
				ChannelID:   ch.ID,
				UnreadCount: ch.UnreadCount,
			}
		}
		PrintJSON(out)
	case Simple:
		for _, ch := range channels {
			fmt.Printf("%s (%d)\n", ch.Name, ch.UnreadCount)
		}
	default:
		if countOnly {
			total := 0
			for _, ch := range channels {
				fmt.Printf("%s: %d\n", ch.Name, ch.UnreadCount)
				total += ch.UnreadCount
			}
			bold := color.New(color.Bold)
			bold.Printf("Total: %d unread messages\n", total)
			return
		}
		bold := color.New(color.Bold)
		bold.Printf("%-17s%-8s%s\n", "Channel", "Unread", "Last Message")
		fmt.Println(Separator(50))
		for _, ch := range channels {
			fmt.Printf("%-17s%-8d%s\n", ch.Name, ch.UnreadCount, ch.LastRead)
		}
	}
}
