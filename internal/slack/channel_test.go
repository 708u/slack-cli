package slack

import (
	"net/http"
	"testing"
)

func TestListChannels(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("conversations.list", slackResponse{
		"ok": true,
		"channels": []any{
			slackChannel("CTEST00001", "general"),
			slackChannel("CTEST00002", "random"),
		},
		"response_metadata": map[string]any{"next_cursor": ""},
	})

	client := mock.newTestClient()
	channels, err := client.ListChannels(ListChannelsOptions{
		Types:           "public_channel",
		ExcludeArchived: true,
		Limit:           100,
	})
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
	if channels[0].ID != "CTEST00001" || channels[0].Name != "general" {
		t.Errorf("unexpected channel[0]: %+v", channels[0])
	}
	if channels[1].ID != "CTEST00002" || channels[1].Name != "random" {
		t.Errorf("unexpected channel[1]: %+v", channels[1])
	}
}

func TestListChannels_Pagination(t *testing.T) {
	mock := newMockSlack(t)
	callCount := 0
	mock.handle("conversations.list", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			writeJSON(w, slackResponse{
				"ok":       true,
				"channels": []any{slackChannel("CTEST00001", "general")},
				"response_metadata": map[string]any{
					"next_cursor": "cursor_page2",
				},
			})
			return
		}
		writeJSON(w, slackResponse{
			"ok":                true,
			"channels":          []any{slackChannel("CTEST00002", "random")},
			"response_metadata": map[string]any{"next_cursor": ""},
		})
	})

	client := mock.newTestClient()
	channels, err := client.ListChannels(ListChannelsOptions{
		Types: "public_channel",
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels from pagination, got %d", len(channels))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestGetChannelDetail(t *testing.T) {
	mock := newMockSlack(t)

	ch := slackChannel("CTEST00001", "general")
	ch["num_members"] = 42
	ch["topic"] = map[string]any{"value": "General chat", "creator": "U001", "last_set": 1600000000}
	ch["purpose"] = map[string]any{"value": "For general discussion", "creator": "U001", "last_set": 1600000000}

	mock.handleJSON("conversations.info", slackResponse{
		"ok":      true,
		"channel": ch,
	})

	client := mock.newTestClient()
	detail, err := client.GetChannelDetail("CTEST00001")
	if err != nil {
		t.Fatalf("GetChannelDetail: %v", err)
	}
	if detail.ID != "CTEST00001" {
		t.Errorf("expected ID C001, got %s", detail.ID)
	}
	if detail.NumMembers != 42 {
		t.Errorf("expected 42 members, got %d", detail.NumMembers)
	}
	if detail.Topic.Value != "General chat" {
		t.Errorf("unexpected topic: %s", detail.Topic.Value)
	}
}

func TestResolveChannelID_AlreadyID(t *testing.T) {
	mock := newMockSlack(t)
	client := mock.newTestClient()

	// Channel ID should be returned as-is without API calls
	id, err := client.ResolveChannelID("C0123456789")
	if err != nil {
		t.Fatalf("ResolveChannelID: %v", err)
	}
	if id != "C0123456789" {
		t.Errorf("expected C0123456789, got %s", id)
	}
}

func TestResolveChannelID_ByName(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("conversations.list", slackResponse{
		"ok": true,
		"channels": []any{
			slackChannel("CTEST00001", "general"),
			slackChannel("CTEST00002", "random"),
		},
		"response_metadata": map[string]any{"next_cursor": ""},
	})

	client := mock.newTestClient()
	id, err := client.ResolveChannelID("random")
	if err != nil {
		t.Fatalf("ResolveChannelID: %v", err)
	}
	if id != "CTEST00002" {
		t.Errorf("expected C002, got %s", id)
	}
}

func TestResolveChannelID_NotFound(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("conversations.list", slackResponse{
		"ok":                true,
		"channels":          []any{slackChannel("CTEST00001", "general")},
		"response_metadata": map[string]any{"next_cursor": ""},
	})

	client := mock.newTestClient()
	_, err := client.ResolveChannelID("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent channel")
	}
}

func TestSetTopic(t *testing.T) {
	mock := newMockSlack(t)

	var receivedTopic string
	mock.handle("conversations.setTopic", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		receivedTopic = r.FormValue("topic")
		writeJSON(w, slackResponse{
			"ok":      true,
			"channel": slackChannel("CTEST00001", "general"),
		})
	})

	client := mock.newTestClient()
	err := client.SetTopic("CTEST00001", "New topic")
	if err != nil {
		t.Fatalf("SetTopic: %v", err)
	}
	if receivedTopic != "New topic" {
		t.Errorf("expected topic 'New topic', got %q", receivedTopic)
	}
}

func TestSetPurpose(t *testing.T) {
	mock := newMockSlack(t)

	var receivedPurpose string
	mock.handle("conversations.setPurpose", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		receivedPurpose = r.FormValue("purpose")
		writeJSON(w, slackResponse{
			"ok":      true,
			"channel": slackChannel("CTEST00001", "general"),
		})
	})

	client := mock.newTestClient()
	err := client.SetPurpose("CTEST00001", "New purpose")
	if err != nil {
		t.Fatalf("SetPurpose: %v", err)
	}
	if receivedPurpose != "New purpose" {
		t.Errorf("expected purpose 'New purpose', got %q", receivedPurpose)
	}
}

func TestJoinChannel(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("conversations.join", slackResponse{
		"ok":      true,
		"channel": slackChannel("CTEST00001", "general"),
	})

	client := mock.newTestClient()
	if err := client.JoinChannel("CTEST00001"); err != nil {
		t.Fatalf("JoinChannel: %v", err)
	}
}

func TestLeaveChannel(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("conversations.leave", slackResponse{
		"ok": true,
	})

	client := mock.newTestClient()
	if err := client.LeaveChannel("CTEST00001"); err != nil {
		t.Fatalf("LeaveChannel: %v", err)
	}
}

func TestInviteToChannel(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("conversations.invite", slackResponse{
		"ok":      true,
		"channel": slackChannel("CTEST00001", "general"),
	})

	client := mock.newTestClient()
	if err := client.InviteToChannel("CTEST00001", []string{"U001", "U002"}, false); err != nil {
		t.Fatalf("InviteToChannel: %v", err)
	}
}

func TestGetChannelMembers(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("conversations.members", slackResponse{
		"ok":                true,
		"members":           []any{"U001", "U002", "U003"},
		"response_metadata": map[string]any{"next_cursor": ""},
	})

	client := mock.newTestClient()
	result, err := client.GetChannelMembers("CTEST00001", nil)
	if err != nil {
		t.Fatalf("GetChannelMembers: %v", err)
	}
	if len(result.Members) != 3 {
		t.Fatalf("expected 3 members, got %d", len(result.Members))
	}
	if result.Members[0] != "U001" {
		t.Errorf("expected first member U001, got %s", result.Members[0])
	}
}
