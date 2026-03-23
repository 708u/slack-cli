package slack

import (
	"testing"
)

func slackUserGroup(id, name, handle, description string, userCount int) map[string]any {
	return map[string]any{
		"id":           id,
		"team_id":      "T0001",
		"is_usergroup": true,
		"name":         name,
		"handle":       handle,
		"description":  description,
		"is_external":  false,
		"date_create":  1600000000,
		"date_update":  1600000000,
		"date_delete":  0,
		"auto_type":    "",
		"created_by":   "U001",
		"updated_by":   "U001",
		"deleted_by":   "",
		"user_count":   userCount,
		"users":        []string{},
		"prefs": map[string]any{
			"channels": []string{},
			"groups":   []string{},
		},
	}
}

func TestListUserGroups(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("usergroups.list", slackResponse{
		"ok": true,
		"usergroups": []any{
			slackUserGroup("S001", "engineering", "eng", "Engineering team", 5),
			slackUserGroup("S002", "design", "design", "Design team", 3),
		},
	})

	client := mock.newTestClient()
	groups, err := client.ListUserGroups(ListUserGroupsOptions{
		IncludeCount: true,
	})
	if err != nil {
		t.Fatalf("ListUserGroups: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Handle != "eng" {
		t.Errorf("expected handle eng, got %s", groups[0].Handle)
	}
	if groups[0].UserCount != 5 {
		t.Errorf("expected user_count 5, got %d", groups[0].UserCount)
	}
	if groups[1].Name != "design" {
		t.Errorf("expected name design, got %s", groups[1].Name)
	}
}

func TestListUserGroups_Empty(t *testing.T) {
	mock := newMockSlack(t)
	mock.handleJSON("usergroups.list", slackResponse{
		"ok":         true,
		"usergroups": []any{},
	})

	client := mock.newTestClient()
	groups, err := client.ListUserGroups(ListUserGroupsOptions{})
	if err != nil {
		t.Fatalf("ListUserGroups: %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
}
