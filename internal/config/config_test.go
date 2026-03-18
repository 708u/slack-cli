package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) (string, *ProfileConfigManager) {
	t.Helper()
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	keyDir := filepath.Join(dir, "keys")

	cs := NewTokenCryptoService(
		WithKeyFilePath(filepath.Join(keyDir, "master.key")),
	)
	mgr := NewProfileConfigManager(
		WithConfigDir(configDir),
		WithCryptoService(cs),
	)
	return dir, mgr
}

func TestSetAndGetBotToken(t *testing.T) {
	_, mgr := setupTestDir(t)

	kind, err := mgr.SetToken("xoxb-test-token", "")
	if err != nil {
		t.Fatalf("SetToken: %v", err)
	}
	if kind != TokenKindBot {
		t.Errorf("expected Bot kind, got %s", kind)
	}

	cfg, err := mgr.GetConfig("")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if cfg.BotToken != "xoxb-test-token" {
		t.Errorf("expected bot token xoxb-test-token, got %s", cfg.BotToken)
	}
	if cfg.UserToken != "" {
		t.Errorf("expected empty user token, got %s", cfg.UserToken)
	}
}

func TestSetAndGetUserToken(t *testing.T) {
	_, mgr := setupTestDir(t)

	kind, err := mgr.SetToken("xoxp-user-token", "")
	if err != nil {
		t.Fatalf("SetToken: %v", err)
	}
	if kind != TokenKindUser {
		t.Errorf("expected User kind, got %s", kind)
	}

	cfg, err := mgr.GetConfig("")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.UserToken != "xoxp-user-token" {
		t.Errorf("expected user token xoxp-user-token, got %s", cfg.UserToken)
	}
}

func TestSetBothTokens(t *testing.T) {
	_, mgr := setupTestDir(t)

	if _, err := mgr.SetToken("xoxb-bot-token", ""); err != nil {
		t.Fatalf("SetToken bot: %v", err)
	}
	if _, err := mgr.SetToken("xoxp-user-token", ""); err != nil {
		t.Fatalf("SetToken user: %v", err)
	}

	cfg, err := mgr.GetConfig("")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg.BotToken != "xoxb-bot-token" {
		t.Errorf("expected bot token, got %s", cfg.BotToken)
	}
	if cfg.UserToken != "xoxp-user-token" {
		t.Errorf("expected user token, got %s", cfg.UserToken)
	}
}

func TestSetAndGetToken_NamedProfile(t *testing.T) {
	_, mgr := setupTestDir(t)

	if _, err := mgr.SetToken("xoxb-work-token", "work"); err != nil {
		t.Fatalf("SetToken: %v", err)
	}

	cfg, err := mgr.GetConfig("work")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if cfg.BotToken != "xoxb-work-token" {
		t.Errorf("expected token xoxb-work-token, got %s", cfg.BotToken)
	}
}

func TestGetConfig_NotFound(t *testing.T) {
	_, mgr := setupTestDir(t)

	cfg, err := mgr.GetConfig("nonexistent")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil for nonexistent profile, got %+v", cfg)
	}
}

func TestListProfiles(t *testing.T) {
	_, mgr := setupTestDir(t)

	mgr.SetToken("xoxb-token1", "default")
	mgr.SetToken("xoxb-token2", "work")

	profiles, err := mgr.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	hasDefault := false
	for _, p := range profiles {
		if p.Name == "default" && p.IsDefault {
			hasDefault = true
		}
	}
	if !hasDefault {
		t.Error("expected 'default' profile to be marked as default")
	}
}

func TestUseProfile(t *testing.T) {
	_, mgr := setupTestDir(t)

	mgr.SetToken("xoxb-token1", "default")
	mgr.SetToken("xoxb-token2", "work")

	if err := mgr.UseProfile("work"); err != nil {
		t.Fatalf("UseProfile: %v", err)
	}

	current, err := mgr.GetCurrentProfile()
	if err != nil {
		t.Fatalf("GetCurrentProfile: %v", err)
	}
	if current != "work" {
		t.Errorf("expected current profile 'work', got %s", current)
	}
}

func TestUseProfile_NotExist(t *testing.T) {
	_, mgr := setupTestDir(t)

	err := mgr.UseProfile("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent profile")
	}
}

func TestClearConfig(t *testing.T) {
	_, mgr := setupTestDir(t)

	mgr.SetToken("xoxb-token1", "default")
	mgr.SetToken("xoxb-token2", "work")

	if err := mgr.ClearConfig("work"); err != nil {
		t.Fatalf("ClearConfig: %v", err)
	}

	profiles, _ := mgr.ListProfiles()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile after clear, got %d", len(profiles))
	}
}

func TestClearConfig_LastProfile(t *testing.T) {
	dir, mgr := setupTestDir(t)

	mgr.SetToken("xoxb-token1", "default")
	if err := mgr.ClearConfig("default"); err != nil {
		t.Fatalf("ClearConfig: %v", err)
	}

	configFile := filepath.Join(dir, "config", "config.json")
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Error("expected config file to be removed after clearing last profile")
	}
}

func TestMaskToken(t *testing.T) {
	_, mgr := setupTestDir(t)

	tests := []struct {
		token    string
		expected string
	}{
		{"xoxb-123456789-abcdef", "xoxb-****-****-cdef"},
		{"short", "****"},
		{"exactlynine", "exac-****-****-nine"},
	}
	for _, tt := range tests {
		got := mgr.MaskToken(tt.token)
		if got != tt.expected {
			t.Errorf("MaskToken(%q) = %q, want %q", tt.token, got, tt.expected)
		}
	}
}

func TestGetCurrentProfile_Default(t *testing.T) {
	_, mgr := setupTestDir(t)

	current, err := mgr.GetCurrentProfile()
	if err != nil {
		t.Fatalf("GetCurrentProfile: %v", err)
	}
	if current != "default" {
		t.Errorf("expected 'default', got %s", current)
	}
}
