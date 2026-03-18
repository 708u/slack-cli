package slack

import (
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// Reaction tests
// ---------------------------------------------------------------------------

func TestAddReaction(t *testing.T) {
	mock := newMockSlack(t)

	var receivedName string
	mock.handle("reactions.add", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		receivedName = r.FormValue("name")
		writeJSON(w, slackResponse{"ok": true})
	})

	client := mock.newTestClient()
	if err := client.AddReaction("CTEST00001", "1234567890.123456", ":thumbsup:"); err != nil {
		t.Fatalf("AddReaction: %v", err)
	}
	if receivedName != "thumbsup" {
		t.Errorf("expected emoji 'thumbsup' (colons stripped), got %q", receivedName)
	}
}

func TestRemoveReaction(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("reactions.remove", slackResponse{"ok": true})

	client := mock.newTestClient()
	if err := client.RemoveReaction("CTEST00001", "1234567890.123456", "thumbsup"); err != nil {
		t.Fatalf("RemoveReaction: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Pin tests
// ---------------------------------------------------------------------------

func TestAddPin(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("pins.add", slackResponse{"ok": true})

	client := mock.newTestClient()
	if err := client.AddPin("CTEST00001", "1234567890.123456"); err != nil {
		t.Fatalf("AddPin: %v", err)
	}
}

func TestRemovePin(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("pins.remove", slackResponse{"ok": true})

	client := mock.newTestClient()
	if err := client.RemovePin("CTEST00001", "1234567890.123456"); err != nil {
		t.Fatalf("RemovePin: %v", err)
	}
}

func TestListPins(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("pins.list", slackResponse{
		"ok": true,
		"items": []any{
			map[string]any{
				"type": "message",
				"message": map[string]any{
					"text": "pinned msg",
					"user": "U001",
					"ts":   "1234567890.123456",
				},
			},
		},
	})

	client := mock.newTestClient()
	items, err := client.ListPins("CTEST00001")
	if err != nil {
		t.Fatalf("ListPins: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 pinned item, got %d", len(items))
	}
	if items[0].Message.Text != "pinned msg" {
		t.Errorf("unexpected pin text: %s", items[0].Message.Text)
	}
}

// ---------------------------------------------------------------------------
// Star tests
// ---------------------------------------------------------------------------

func TestAddStar(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("stars.add", slackResponse{"ok": true})

	client := mock.newTestClient()
	if err := client.AddStar("CTEST00001", "1234567890.123456"); err != nil {
		t.Fatalf("AddStar: %v", err)
	}
}

func TestRemoveStar(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("stars.remove", slackResponse{"ok": true})

	client := mock.newTestClient()
	if err := client.RemoveStar("CTEST00001", "1234567890.123456"); err != nil {
		t.Fatalf("RemoveStar: %v", err)
	}
}

func TestListStars(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("stars.list", slackResponse{
		"ok": true,
		"items": []any{
			map[string]any{
				"type":    "message",
				"channel": "CTEST00001",
				"message": map[string]any{
					"text": "starred msg",
					"ts":   "1234567890.123456",
				},
				"date_create": 1700000000,
			},
		},
	})

	client := mock.newTestClient()
	result, err := client.ListStars(100)
	if err != nil {
		t.Fatalf("ListStars: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 star, got %d", len(result.Items))
	}
	if result.Items[0].Message.Text != "starred msg" {
		t.Errorf("unexpected star text: %s", result.Items[0].Message.Text)
	}
}

// ---------------------------------------------------------------------------
// Reminder tests
// ---------------------------------------------------------------------------

func TestAddReminder(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("reminders.add", slackResponse{
		"ok": true,
		"reminder": map[string]any{
			"id":          "Rm001",
			"text":        "do the thing",
			"time":        1700000000,
			"complete_ts": 0,
			"recurring":   false,
		},
	})

	client := mock.newTestClient()
	rem, err := client.AddReminder("do the thing", 1700000000)
	if err != nil {
		t.Fatalf("AddReminder: %v", err)
	}
	if rem.ID != "Rm001" {
		t.Errorf("expected ID Rm001, got %s", rem.ID)
	}
	if rem.Text != "do the thing" {
		t.Errorf("unexpected text: %s", rem.Text)
	}
}

func TestListReminders(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("reminders.list", slackResponse{
		"ok": true,
		"reminders": []any{
			map[string]any{
				"id":          "Rm001",
				"text":        "do the thing",
				"time":        1700000000,
				"complete_ts": 0,
				"recurring":   false,
			},
		},
	})

	client := mock.newTestClient()
	rems, err := client.ListReminders()
	if err != nil {
		t.Fatalf("ListReminders: %v", err)
	}
	if len(rems) != 1 {
		t.Fatalf("expected 1 reminder, got %d", len(rems))
	}
}

func TestDeleteReminder(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("reminders.delete", slackResponse{"ok": true})

	client := mock.newTestClient()
	if err := client.DeleteReminder("Rm001"); err != nil {
		t.Fatalf("DeleteReminder: %v", err)
	}
}

func TestCompleteReminder(t *testing.T) {
	mock := newMockSlack(t)
	// CompleteReminder uses raw HTTP, not slack-go
	mock.handleJSON("reminders.complete", slackResponse{"ok": true})

	client := mock.newTestClient()
	if err := client.CompleteReminder("Rm001"); err != nil {
		t.Fatalf("CompleteReminder: %v", err)
	}
}

// ---------------------------------------------------------------------------
// File upload tests
// ---------------------------------------------------------------------------

func TestUploadFile_Content(t *testing.T) {
	mock := newMockSlack(t)

	// slack-go v2 upload flow: getUploadURLExternal → PUT to URL → completeUploadExternal
	mock.handle("files.getUploadURLExternal", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, slackResponse{
			"ok":         true,
			"upload_url": mock.server.URL + "/upload/F001",
			"file_id":    "F001",
		})
	})
	// Handle the actual upload PUT/POST
	mock.mux.HandleFunc("/upload/F001", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mock.handleJSON("files.completeUploadExternal", slackResponse{
		"ok":    true,
		"files": []any{map[string]any{"id": "F001"}},
	})

	client := mock.newTestClient()
	err := client.UploadFile(UploadFileOptions{
		Channel:  "CTEST00001",
		Content:  "code snippet here",
		Filename: "snippet.py",
		Title:    "My Snippet",
	})
	if err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Canvas tests
// ---------------------------------------------------------------------------

func TestReadCanvas(t *testing.T) {
	mock := newMockSlack(t)
	// ReadCanvas uses raw HTTP POST
	mock.handleJSON("canvases.sections.lookup", slackResponse{
		"ok": true,
		"sections": []any{
			map[string]any{
				"id": "sec001",
				"elements": []any{
					map[string]any{
						"type": "text",
						"text": "Hello Canvas",
					},
				},
			},
		},
	})

	client := mock.newTestClient()
	sections, err := client.ReadCanvas("canvas123")
	if err != nil {
		t.Fatalf("ReadCanvas: %v", err)
	}
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].ID != "sec001" {
		t.Errorf("unexpected section ID: %s", sections[0].ID)
	}
	if len(sections[0].Elements) != 1 || sections[0].Elements[0].Text != "Hello Canvas" {
		t.Errorf("unexpected elements: %+v", sections[0].Elements)
	}
}

func TestListCanvases(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("files.list", slackResponse{
		"ok": true,
		"files": []any{
			map[string]any{
				"id":       "F001",
				"name":     "My Canvas",
				"created":  1700000000,
				"filetype": "spaces",
			},
		},
	})

	client := mock.newTestClient()
	canvases, err := client.ListCanvases("CTEST00001")
	if err != nil {
		t.Fatalf("ListCanvases: %v", err)
	}
	if len(canvases) != 1 {
		t.Fatalf("expected 1 canvas, got %d", len(canvases))
	}
	if canvases[0].Name != "My Canvas" {
		t.Errorf("unexpected canvas name: %s", canvases[0].Name)
	}
}
