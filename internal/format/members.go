package format

import (
	"fmt"

	"github.com/fatih/color"
)

// MemberInfo holds the fields needed to display a channel member row.
type MemberInfo struct {
	ID       string
	Name     string
	RealName string
}

// FormatMembers prints a list of channel members in the requested format.
func FormatMembers(members []MemberInfo, f Format) {
	switch f {
	case JSON:
		type jsonMember struct {
			ID       string `json:"id"`
			Name     string `json:"name,omitempty"`
			RealName string `json:"real_name,omitempty"`
		}
		out := make([]jsonMember, len(members))
		for i, m := range members {
			out[i] = jsonMember{
				ID:       m.ID,
				Name:     m.Name,
				RealName: m.RealName,
			}
		}
		PrintJSON(out)
	case Simple:
		for _, m := range members {
			if m.Name != "" {
				fmt.Printf("%s\t%s\t%s\n", m.ID, m.Name, m.RealName)
			} else {
				fmt.Println(m.ID)
			}
		}
	default:
		bold := color.New(color.Bold)
		bold.Printf("%-13s%-18s%s\n", "ID", "Name", "Real Name")
		fmt.Println(Separator(50))
		for _, m := range members {
			fmt.Printf("%-13s%-18s%s\n",
				m.ID,
				truncate(m.Name, 16),
				truncate(m.RealName, 20),
			)
		}
	}
}
