package format

import (
	"fmt"

	"github.com/fatih/color"
)

// UserInfo holds the fields needed to display a user list row.
type UserInfo struct {
	ID          string
	Name        string
	RealName    string
	DisplayName string
	IsBot       bool
	IsAdmin     bool
}

// FormatUserList prints a list of users in the requested format.
func FormatUserList(users []UserInfo, f Format) {
	switch f {
	case JSON:
		type jsonUser struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			RealName    string `json:"real_name"`
			DisplayName string `json:"display_name"`
			IsBot       bool   `json:"is_bot"`
			IsAdmin     bool   `json:"is_admin"`
		}
		out := make([]jsonUser, len(users))
		for i, u := range users {
			out[i] = jsonUser{
				ID:          u.ID,
				Name:        u.Name,
				RealName:    u.RealName,
				DisplayName: u.DisplayName,
				IsBot:       u.IsBot,
				IsAdmin:     u.IsAdmin,
			}
		}
		PrintJSON(out)
	case Simple:
		for _, u := range users {
			fmt.Printf("%s\t%s\t%s\t%s\n", u.ID, u.Name, u.RealName, u.DisplayName)
		}
	default:
		bold := color.New(color.Bold)
		bold.Printf("%-12s%-20s%-20s%-20s%-6s%-6s\n", "ID", "Name", "Real Name", "Display Name", "Bot", "Admin")
		fmt.Println(Separator(84))
		for _, u := range users {
			fmt.Printf("%-12s%-20s%-20s%-20s%-6s%-6s\n",
				u.ID, u.Name, u.RealName, u.DisplayName,
				boolYesNo(u.IsBot), boolYesNo(u.IsAdmin))
		}
	}
}

// UserDetailInfo holds the fields needed to display detailed user info.
type UserDetailInfo struct {
	ID          string
	Name        string
	RealName    string
	Email       string
	Title       string
	DisplayName string
	StatusText  string
	StatusEmoji string
	TZ          string
	TZLabel     string
	IsAdmin     bool
	IsBot       bool
	Deleted     bool
}

// FormatUserInfo prints detailed information about a single user.
func FormatUserInfo(user UserDetailInfo, f Format) {
	switch f {
	case JSON:
		out := struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			RealName    string `json:"real_name"`
			Email       string `json:"email,omitempty"`
			Title       string `json:"title,omitempty"`
			DisplayName string `json:"display_name,omitempty"`
			StatusText  string `json:"status_text,omitempty"`
			StatusEmoji string `json:"status_emoji,omitempty"`
			TZ          string `json:"tz,omitempty"`
			TZLabel     string `json:"tz_label,omitempty"`
			IsAdmin     bool   `json:"is_admin"`
			IsBot       bool   `json:"is_bot"`
			Deleted     bool   `json:"deleted"`
		}{
			ID:          user.ID,
			Name:        user.Name,
			RealName:    user.RealName,
			Email:       user.Email,
			Title:       user.Title,
			DisplayName: user.DisplayName,
			StatusText:  user.StatusText,
			StatusEmoji: user.StatusEmoji,
			TZ:          user.TZ,
			TZLabel:     user.TZLabel,
			IsAdmin:     user.IsAdmin,
			IsBot:       user.IsBot,
			Deleted:     user.Deleted,
		}
		PrintJSON(out)
	case Simple:
		fmt.Printf("%s\t%s\t%s\n", user.ID, user.Name, user.RealName)
	default:
		bold := color.New(color.Bold)
		gray := color.New(color.FgHiBlack)

		bold.Printf("\nUser Info: %s\n", user.Name)
		fmt.Println()
		fmt.Printf("  %s %s\n", gray.Sprint("ID:"), user.ID)
		fmt.Printf("  %s %s\n", gray.Sprint("Name:"), user.Name)
		fmt.Printf("  %s %s\n", gray.Sprint("Real Name:"), user.RealName)
		if user.DisplayName != "" {
			fmt.Printf("  %s %s\n", gray.Sprint("Display:"), user.DisplayName)
		}
		if user.Email != "" {
			fmt.Printf("  %s %s\n", gray.Sprint("Email:"), user.Email)
		}
		if user.Title != "" {
			fmt.Printf("  %s %s\n", gray.Sprint("Title:"), user.Title)
		}
		if user.StatusText != "" {
			status := user.StatusText
			if user.StatusEmoji != "" {
				status = user.StatusEmoji + " " + status
			}
			fmt.Printf("  %s %s\n", gray.Sprint("Status:"), status)
		}
		if user.TZLabel != "" {
			fmt.Printf("  %s %s\n", gray.Sprint("Timezone:"), user.TZLabel)
		}
		fmt.Printf("  %s %s\n", gray.Sprint("Admin:"), boolYesNo(user.IsAdmin))
		fmt.Printf("  %s %s\n", gray.Sprint("Bot:"), boolYesNo(user.IsBot))
		fmt.Printf("  %s %s\n", gray.Sprint("Deleted:"), boolYesNo(user.Deleted))
		fmt.Println()
	}
}

// FormatPresence prints the presence status of a user.
func FormatPresence(userID string, presence string, f Format) {
	switch f {
	case JSON:
		out := struct {
			UserID   string `json:"user_id"`
			Presence string `json:"presence"`
		}{
			UserID:   userID,
			Presence: presence,
		}
		PrintJSON(out)
	case Simple:
		fmt.Printf("%s\t%s\n", userID, presence)
	default:
		gray := color.New(color.FgHiBlack)
		var presenceColor *color.Color
		if presence == "active" {
			presenceColor = color.New(color.FgGreen)
		} else {
			presenceColor = color.New(color.FgYellow)
		}
		fmt.Printf("%s %s\n", gray.Sprint("Presence:"), presenceColor.Sprint(presence))
	}
}
