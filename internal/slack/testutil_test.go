package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	slackgo "github.com/slack-go/slack"
)

// slackResponse is a helper to build Slack API JSON responses.
type slackResponse map[string]any

func (r slackResponse) withOK() slackResponse {
	r["ok"] = true
	return r
}

// mockSlack holds a test HTTP server that mimics the Slack API.
// Individual handlers are registered by API method name.
type mockSlack struct {
	server   *httptest.Server
	mux      *http.ServeMux
	handlers map[string]http.HandlerFunc
}

// newMockSlack creates a new mock server. Call close() when done.
func newMockSlack(t *testing.T) *mockSlack {
	t.Helper()

	m := &mockSlack{
		handlers: make(map[string]http.HandlerFunc),
	}

	mux := http.NewServeMux()
	// Catch-all pattern: "/api/<method>"
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Extract method name from path: "/api/chat.postMessage" -> "chat.postMessage"
		method := r.URL.Path[len("/api/"):]

		if handler, ok := m.handlers[method]; ok {
			handler(w, r)
			return
		}

		// Default: return ok with empty result
		writeJSON(w, slackResponse{"ok": true})
	})

	m.mux = mux
	m.server = httptest.NewServer(mux)
	t.Cleanup(m.server.Close)

	return m
}

// handle registers a handler for the given Slack API method.
func (m *mockSlack) handle(method string, handler http.HandlerFunc) {
	m.handlers[method] = handler
}

// handleJSON registers a handler that returns the given response as JSON.
func (m *mockSlack) handleJSON(method string, resp slackResponse) {
	m.handlers[method] = func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, resp)
	}
}

// client creates a slack-go Client pointing at the mock server.
func (m *mockSlack) client() *slackgo.Client {
	return slackgo.New(
		"xoxb-test-token",
		slackgo.OptionAPIURL(m.server.URL+"/api/"),
	)
}

// newTestClient creates a full Client backed by the mock server.
func (m *mockSlack) newTestClient() *Client {
	api := m.client()
	c := &Client{api: api}
	c.channelOps = newChannelOps(api)
	c.messageOps = newMessageOps(api, c.channelOps)
	c.userOps = newUserOps(api)
	c.searchOps = newSearchOps(api)
	c.reactionOps = newReactionOps(api, c.channelOps)
	c.pinOps = newPinOps(api, c.channelOps)
	c.reminderOps = newReminderOps(api, "xoxb-test-token")
	c.reminderOps.baseURL = m.server.URL + "/api/"
	c.starOps = newStarOps(api)
	c.fileOps = newFileOps(api, c.channelOps)
	c.canvasOps = newCanvasOps(api, "xoxb-test-token", c.channelOps)
	c.canvasOps.baseURL = m.server.URL + "/api/"

	return c
}

// writeJSON writes v as JSON to w.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// slackChannel is a helper to build channel JSON for mock responses.
func slackChannel(id, name string) map[string]any {
	return map[string]any{
		"id":                   id,
		"name":                 name,
		"is_channel":           true,
		"is_group":             false,
		"is_im":                false,
		"is_mpim":              false,
		"is_private":           false,
		"created":              1600000000,
		"is_archived":          false,
		"is_general":           false,
		"name_normalized":      name,
		"is_shared":            false,
		"is_ext_shared":        false,
		"is_org_shared":        false,
		"is_member":            true,
		"num_members":          5,
		"unread_count":         0,
		"unread_count_display": 0,
		"topic":                map[string]any{"value": "", "creator": "", "last_set": 0},
		"purpose":              map[string]any{"value": "", "creator": "", "last_set": 0},
	}
}

// slackMessage is a helper to build message JSON for mock responses.
func slackMessage(user, text, ts string) map[string]any {
	return map[string]any{
		"type": "message",
		"user": user,
		"text": text,
		"ts":   ts,
	}
}

// slackUser is a helper to build user JSON for mock responses.
func slackUser(id, name, realName, email string) map[string]any {
	return map[string]any{
		"id":        id,
		"name":      name,
		"real_name": realName,
		"deleted":   false,
		"is_bot":    false,
		"is_admin":  false,
		"tz":        "America/Los_Angeles",
		"tz_label":  "Pacific Standard Time",
		"profile": map[string]any{
			"email":        email,
			"display_name": name,
			"title":        "",
			"status_text":  "",
			"status_emoji": "",
		},
	}
}
