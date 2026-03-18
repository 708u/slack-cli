package slack

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	slackgo "github.com/slack-go/slack"
)

// ReminderOps groups reminder-related API calls. It stores the raw
// token because certain reminder endpoints (e.g. reminders.complete)
// are not available in the slack-go library and require direct HTTP
// calls.
type ReminderOps struct {
	api     *slackgo.Client
	token   string
	baseURL string // defaults to slackgo.APIURL
}

func newReminderOps(api *slackgo.Client, token string) *ReminderOps {
	return &ReminderOps{
		api:     api,
		token:   token,
		baseURL: slackgo.APIURL,
	}
}

// AddReminder creates a reminder for the authenticated user at the
// specified Unix epoch time.
func (r *ReminderOps) AddReminder(text string, time int64) (*Reminder, error) {
	rem, err := r.api.AddUserReminder("me", text, strconv.FormatInt(time, 10))
	if err != nil {
		return nil, fmt.Errorf("add reminder: %w", err)
	}

	return &Reminder{
		ID:         rem.ID,
		Text:       rem.Text,
		Time:       int64(rem.Time),
		CompleteTS: int64(rem.CompleteTS),
		Recurring:  rem.Recurring,
	}, nil
}

// ListReminders returns all reminders for the authenticated user.
func (r *ReminderOps) ListReminders() ([]Reminder, error) {
	rems, err := r.api.ListReminders()
	if err != nil {
		return nil, wrapSlackError("list reminders", err)
	}

	out := make([]Reminder, 0, len(rems))
	for _, rem := range rems {
		out = append(out, Reminder{
			ID:         rem.ID,
			Text:       rem.Text,
			Time:       int64(rem.Time),
			CompleteTS: int64(rem.CompleteTS),
			Recurring:  rem.Recurring,
		})
	}
	return out, nil
}

// DeleteReminder deletes a reminder by its ID.
func (r *ReminderOps) DeleteReminder(id string) error {
	if err := r.api.DeleteReminder(id); err != nil {
		return wrapSlackError("delete reminder", err)
	}
	return nil
}

// completeReminderResponse is the Slack API response for
// reminders.complete.
type completeReminderResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// CompleteReminder marks a reminder as complete. This method uses a
// direct HTTP call because the slack-go library does not expose
// reminders.complete.
func (r *ReminderOps) CompleteReminder(id string) error {
	vals := url.Values{
		"token":    {r.token},
		"reminder": {id},
	}

	resp, err := http.PostForm(r.baseURL+"reminders.complete", vals)
	if err != nil {
		return fmt.Errorf("complete reminder: %w", err)
	}
	defer resp.Body.Close()

	var result completeReminderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("complete reminder: decode response: %w", err)
	}
	if !result.OK {
		if result.Error == "not_allowed_token_type" {
			return fmt.Errorf("complete reminder: this command requires a User token (xoxp-), not a Bot token (xoxb-)")
		}
		return fmt.Errorf("complete reminder: %s", result.Error)
	}
	return nil
}
