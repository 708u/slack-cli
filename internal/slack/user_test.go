package slack

import (
	"net/http"
	"testing"
)

func TestListUsers(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("users.list", slackResponse{
		"ok": true,
		"members": []any{
			slackUser("U001", "alice", "Alice Smith", "alice@example.com"),
			slackUser("U002", "bob", "Bob Jones", "bob@example.com"),
		},
		"response_metadata": map[string]any{"next_cursor": ""},
	})

	client := mock.newTestClient()
	users, err := client.ListUsers(10)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "alice" {
		t.Errorf("expected name alice, got %s", users[0].Name)
	}
	if users[0].Profile.Email != "alice@example.com" {
		t.Errorf("unexpected email: %s", users[0].Profile.Email)
	}
}

func TestGetUserInfo(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("users.info", slackResponse{
		"ok":   true,
		"user": slackUser("U001", "alice", "Alice Smith", "alice@example.com"),
	})

	client := mock.newTestClient()
	user, err := client.GetUserInfo("U001")
	if err != nil {
		t.Fatalf("GetUserInfo: %v", err)
	}
	if user.ID != "U001" {
		t.Errorf("expected ID U001, got %s", user.ID)
	}
	if user.RealName != "Alice Smith" {
		t.Errorf("expected real_name Alice Smith, got %s", user.RealName)
	}
}

func TestLookupUserByEmail(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("users.lookupByEmail", slackResponse{
		"ok":   true,
		"user": slackUser("U001", "alice", "Alice Smith", "alice@example.com"),
	})

	client := mock.newTestClient()
	user, err := client.LookupUserByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("LookupUserByEmail: %v", err)
	}
	if user.ID != "U001" {
		t.Errorf("expected ID U001, got %s", user.ID)
	}
}

func TestOpenDMChannel(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("conversations.open", slackResponse{
		"ok": true,
		"channel": map[string]any{
			"id": "D001",
		},
	})

	client := mock.newTestClient()
	chID, err := client.OpenDMChannel("U001")
	if err != nil {
		t.Fatalf("OpenDMChannel: %v", err)
	}
	if chID != "D001" {
		t.Errorf("expected D001, got %s", chID)
	}
}

func TestGetUserPresence(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("users.getPresence", slackResponse{
		"ok":       true,
		"presence": "active",
	})

	client := mock.newTestClient()
	p, err := client.GetUserPresence("U001")
	if err != nil {
		t.Fatalf("GetUserPresence: %v", err)
	}
	if p.Presence != "active" {
		t.Errorf("expected active, got %s", p.Presence)
	}
}

func TestResolveUserIDByName(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("users.list", slackResponse{
		"ok": true,
		"members": []any{
			slackUser("U001", "alice", "Alice Smith", "alice@example.com"),
			slackUser("U002", "bob", "Bob Jones", "bob@example.com"),
		},
		"response_metadata": map[string]any{"next_cursor": ""},
	})

	client := mock.newTestClient()
	id, err := client.ResolveUserIDByName("@bob")
	if err != nil {
		t.Fatalf("ResolveUserIDByName: %v", err)
	}
	if id != "U002" {
		t.Errorf("expected U002, got %s", id)
	}
}

func TestResolveUserIDByName_NotFound(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("users.list", slackResponse{
		"ok":                true,
		"members":           []any{slackUser("U001", "alice", "Alice", "")},
		"response_metadata": map[string]any{"next_cursor": ""},
	})

	client := mock.newTestClient()
	_, err := client.ResolveUserIDByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}

func TestResolveUserID_AlreadyID(t *testing.T) {
	// No mock handlers needed — if the ID pattern matches, no API call
	// should be made. A call to users.list would panic on missing handler.
	mock := newMockSlack(t)
	client := mock.newTestClient()

	id, err := client.ResolveUserID("U0123456789")
	if err != nil {
		t.Fatalf("ResolveUserID: %v", err)
	}
	if id != "U0123456789" {
		t.Errorf("expected U0123456789, got %s", id)
	}
}

func TestResolveUserID_EnterpriseGrid(t *testing.T) {
	mock := newMockSlack(t)
	client := mock.newTestClient()

	id, err := client.ResolveUserID("W0123456789")
	if err != nil {
		t.Fatalf("ResolveUserID: %v", err)
	}
	if id != "W0123456789" {
		t.Errorf("expected W0123456789, got %s", id)
	}
}

func TestResolveUserID_ByName(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("users.list", slackResponse{
		"ok": true,
		"members": []any{
			slackUser("U001", "alice", "Alice Smith", "alice@example.com"),
			slackUser("U002", "bob", "Bob Jones", "bob@example.com"),
		},
		"response_metadata": map[string]any{"next_cursor": ""},
	})

	client := mock.newTestClient()
	id, err := client.ResolveUserID("@bob")
	if err != nil {
		t.Fatalf("ResolveUserID: %v", err)
	}
	if id != "U002" {
		t.Errorf("expected U002, got %s", id)
	}
}

func TestListUsers_WithLimit(t *testing.T) {
	mock := newMockSlack(t)
	mock.handle("users.list", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, slackResponse{
			"ok": true,
			"members": []any{
				slackUser("U001", "alice", "Alice", ""),
				slackUser("U002", "bob", "Bob", ""),
				slackUser("U003", "charlie", "Charlie", ""),
			},
			"response_metadata": map[string]any{"next_cursor": ""},
		})
	})

	client := mock.newTestClient()
	users, err := client.ListUsers(2)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users (limited), got %d", len(users))
	}
}
