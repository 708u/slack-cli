package slack

import (
	"net/http"
	"testing"
)

func TestSendMessage(t *testing.T) {
	mock := newMockSlack(t)

	var receivedChannel, receivedText string
	mock.handle("chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		receivedChannel = r.FormValue("channel")
		receivedText = r.FormValue("text")
		writeJSON(w, slackResponse{
			"ok":      true,
			"channel": receivedChannel,
			"ts":      "1234567890.123456",
		})
	})

	client := mock.newTestClient()
	err := client.messageOps.SendMessage("CTEST00001", "hello world", "")
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if receivedChannel != "CTEST00001" {
		t.Errorf("expected channel C001, got %s", receivedChannel)
	}
	if receivedText != "hello world" {
		t.Errorf("expected text 'hello world', got %s", receivedText)
	}
}

func TestSendMessage_WithThread(t *testing.T) {
	mock := newMockSlack(t)

	var receivedThreadTS string
	mock.handle("chat.postMessage", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		receivedThreadTS = r.FormValue("thread_ts")
		writeJSON(w, slackResponse{
			"ok":      true,
			"channel": "CTEST00001",
			"ts":      "1234567890.123457",
		})
	})

	client := mock.newTestClient()
	err := client.messageOps.SendMessage("CTEST00001", "reply", "1234567890.123456")
	if err != nil {
		t.Fatalf("SendMessage with thread: %v", err)
	}
	if receivedThreadTS != "1234567890.123456" {
		t.Errorf("expected thread_ts, got %q", receivedThreadTS)
	}
}

func TestSendEphemeralMessage(t *testing.T) {
	mock := newMockSlack(t)

	var receivedUser string
	mock.handle("chat.postEphemeral", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		receivedUser = r.FormValue("user")
		writeJSON(w, slackResponse{
			"ok":         true,
			"message_ts": "1234567890.123456",
		})
	})

	client := mock.newTestClient()
	err := client.messageOps.SendEphemeralMessage("CTEST00001", "U001", "secret msg", "")
	if err != nil {
		t.Fatalf("SendEphemeralMessage: %v", err)
	}
	if receivedUser != "U001" {
		t.Errorf("expected user U001, got %s", receivedUser)
	}
}

func TestUpdateMessage(t *testing.T) {
	mock := newMockSlack(t)

	var receivedTS, receivedText string
	mock.handle("chat.update", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		receivedTS = r.FormValue("ts")
		receivedText = r.FormValue("text")
		writeJSON(w, slackResponse{
			"ok":      true,
			"channel": "CTEST00001",
			"ts":      receivedTS,
			"text":    receivedText,
		})
	})

	client := mock.newTestClient()
	err := client.messageOps.UpdateMessage("CTEST00001", "1234567890.123456", "edited")
	if err != nil {
		t.Fatalf("UpdateMessage: %v", err)
	}
	if receivedTS != "1234567890.123456" {
		t.Errorf("expected ts 1234567890.123456, got %s", receivedTS)
	}
	if receivedText != "edited" {
		t.Errorf("expected text 'edited', got %s", receivedText)
	}
}

func TestDeleteMessage(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("chat.delete", slackResponse{
		"ok":      true,
		"channel": "CTEST00001",
		"ts":      "1234567890.123456",
	})

	client := mock.newTestClient()
	err := client.messageOps.DeleteMessage("CTEST00001", "1234567890.123456")
	if err != nil {
		t.Fatalf("DeleteMessage: %v", err)
	}
}

func TestGetHistory(t *testing.T) {
	mock := newMockSlack(t)

	mock.handleJSON("conversations.history", slackResponse{
		"ok": true,
		"messages": []any{
			slackMessage("U001", "hello", "1234567890.100000"),
			slackMessage("U002", "world", "1234567890.200000"),
		},
		"response_metadata": map[string]any{"next_cursor": ""},
	})
	mock.handle("users.info", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		uid := r.FormValue("user")
		name := "user-" + uid
		writeJSON(w, slackResponse{
			"ok":   true,
			"user": slackUser(uid, name, name, uid+"@example.com"),
		})
	})

	client := mock.newTestClient()
	result, err := client.messageOps.GetHistory("CTEST00001", HistoryOptions{Limit: 10})
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(result.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result.Messages))
	}
	if result.Messages[0].Text != "hello" {
		t.Errorf("unexpected first message text: %s", result.Messages[0].Text)
	}
	if _, ok := result.Users["U001"]; !ok {
		t.Error("expected user U001 in resolved users map")
	}
}

func TestGetThreadHistory(t *testing.T) {
	mock := newMockSlack(t)

	mock.handleJSON("conversations.replies", slackResponse{
		"ok": true,
		"messages": []any{
			slackMessage("U001", "parent", "1234567890.100000"),
			slackMessage("U002", "reply1", "1234567890.100001"),
		},
		"response_metadata": map[string]any{"next_cursor": ""},
	})
	mock.handle("users.info", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		uid := r.FormValue("user")
		writeJSON(w, slackResponse{
			"ok":   true,
			"user": slackUser(uid, "user-"+uid, "User "+uid, ""),
		})
	})

	client := mock.newTestClient()
	result, err := client.messageOps.GetThreadHistory("CTEST00001", "1234567890.100000")
	if err != nil {
		t.Fatalf("GetThreadHistory: %v", err)
	}
	if len(result.Messages) != 2 {
		t.Fatalf("expected 2 messages in thread, got %d", len(result.Messages))
	}
}

func TestGetPermalink(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("chat.getPermalink", slackResponse{
		"ok":        true,
		"permalink": "https://slack.com/archives/C001/p1234567890123456",
		"channel":   "CTEST00001",
	})

	client := mock.newTestClient()
	link, err := client.messageOps.GetPermalink("CTEST00001", "1234567890.123456")
	if err != nil {
		t.Fatalf("GetPermalink: %v", err)
	}
	if link != "https://slack.com/archives/C001/p1234567890123456" {
		t.Errorf("unexpected permalink: %s", link)
	}
}

func TestGetPermalinks_Multiple(t *testing.T) {
	mock := newMockSlack(t)
	mock.handle("chat.getPermalink", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		ts := r.FormValue("message_ts")
		writeJSON(w, slackResponse{
			"ok":        true,
			"permalink": "https://slack.com/archives/C001/p" + ts,
			"channel":   "CTEST00001",
		})
	})

	client := mock.newTestClient()
	links, err := client.messageOps.GetPermalinks("CTEST00001", []string{"1234567890.100000", "1234567890.200000"})
	if err != nil {
		t.Fatalf("GetPermalinks: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 permalinks, got %d", len(links))
	}
}

func TestMarkAsRead(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("conversations.mark", slackResponse{"ok": true})

	client := mock.newTestClient()
	if err := client.messageOps.MarkAsRead("CTEST00001"); err != nil {
		t.Fatalf("MarkAsRead: %v", err)
	}
}

func TestListScheduledMessages(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("chat.scheduledMessages.list", slackResponse{
		"ok": true,
		"scheduled_messages": []any{
			map[string]any{
				"id":           "Q001",
				"channel_id":   "CTEST00001",
				"post_at":      1700000000,
				"date_created": 1699000000,
				"text":         "scheduled msg",
			},
		},
	})

	client := mock.newTestClient()
	msgs, err := client.messageOps.ListScheduledMessages("", 50)
	if err != nil {
		t.Fatalf("ListScheduledMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 scheduled message, got %d", len(msgs))
	}
	if msgs[0].ID != "Q001" {
		t.Errorf("expected ID Q001, got %s", msgs[0].ID)
	}
}
