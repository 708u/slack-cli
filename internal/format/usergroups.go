package format

import (
	"fmt"

	"github.com/fatih/color"
)

// UserGroupInfo holds the fields needed to display a user group list row.
type UserGroupInfo struct {
	ID          string
	Name        string
	Handle      string
	Description string
	UserCount   int
	IsExternal  bool
}

// FormatUserGroupList prints a list of user groups in the requested format.
func FormatUserGroupList(groups []UserGroupInfo, f Format) {
	switch f {
	case JSON:
		type jsonUserGroup struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Handle      string `json:"handle"`
			Description string `json:"description"`
			UserCount   int    `json:"user_count"`
			IsExternal  bool   `json:"is_external"`
		}
		out := make([]jsonUserGroup, len(groups))
		for i, g := range groups {
			out[i] = jsonUserGroup{
				ID:          g.ID,
				Name:        g.Name,
				Handle:      g.Handle,
				Description: g.Description,
				UserCount:   g.UserCount,
				IsExternal:  g.IsExternal,
			}
		}
		PrintJSON(out)
	case Simple:
		for _, g := range groups {
			fmt.Printf("%s\t%s\t%s\t%d\n",
				g.ID, g.Handle, g.Name, g.UserCount)
		}
	default:
		const (
			idW     = 14
			nameW   = 22
			handleW = 18
			countW  = 10
			descW   = 30
		)
		bold := color.New(color.Bold)
		bold.Printf("%-*s%-*s%-*s%-*s%s\n",
			idW, "ID",
			nameW, "Name",
			handleW, "Handle",
			countW, "Members",
			"Description")
		fmt.Println(Separator(idW + nameW + handleW + countW + descW))
		for _, g := range groups {
			fmt.Printf("%-*s%-*s%-*s%-*d%s\n",
				idW, g.ID,
				nameW, truncate(g.Name, nameW-2),
				handleW, g.Handle,
				countW, g.UserCount,
				truncate(g.Description, descW-2))
		}
	}
}
