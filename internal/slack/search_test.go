package slack

import (
	"testing"
)

func TestSearchMessages(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("search.messages", slackResponse{
		"ok": true,
		"messages": map[string]any{
			"matches": []any{
				map[string]any{
					"text":      "hello world",
					"user":      "U001",
					"username":  "alice",
					"ts":        "1234567890.100000",
					"permalink": "https://slack.com/archives/CTEST00001/p1234567890100000",
					"channel": map[string]any{
						"id":   "CTEST00001",
						"name": "general",
					},
				},
			},
			"pagination": map[string]any{
				"total_count": 1,
				"page":        1,
				"page_count":  1,
			},
		},
	})

	client := mock.newTestClient()
	result, err := client.SearchMessages("hello", &SearchMessagesOptions{
		Sort:    "score",
		SortDir: "desc",
		Count:   20,
		Page:    1,
	})
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	if result.TotalCount != 1 {
		t.Errorf("expected total_count 1, got %d", result.TotalCount)
	}
	if len(result.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(result.Matches))
	}
	if result.Matches[0].Text != "hello world" {
		t.Errorf("unexpected match text: %s", result.Matches[0].Text)
	}
	if result.Matches[0].Channel.Name != "general" {
		t.Errorf("unexpected channel name: %s", result.Matches[0].Channel.Name)
	}
}

func TestSearchMessages_DefaultOptions(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("search.messages", slackResponse{
		"ok": true,
		"messages": map[string]any{
			"matches":    []any{},
			"pagination": map[string]any{"total_count": 0, "page": 1, "page_count": 0},
		},
	})

	client := mock.newTestClient()
	result, err := client.SearchMessages("test", nil)
	if err != nil {
		t.Fatalf("SearchMessages: %v", err)
	}
	// Verify defaults are applied (no panic, returns valid result)
	if result.Query != "test" {
		t.Errorf("expected query 'test', got %q", result.Query)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected 0 results, got %d", result.TotalCount)
	}
}
