package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	tokenMaskLength    = 4
	tokenMinLength     = 9
	defaultProfileName = "default"
	dirPermission      = 0700
	filePermission     = 0600
	configFileName     = "config.json"
)

// Config holds tokens and metadata for a single profile.
// A profile may have a BotToken, a UserToken, or both.
type Config struct {
	BotToken  string `json:"botToken,omitempty"`
	UserToken string `json:"userToken,omitempty"`
	Timezone  string `json:"timezone,omitempty"`
	UpdatedAt string `json:"updatedAt"`
}

// Profile represents a named configuration profile.
type Profile struct {
	Name      string
	Config    Config
	IsDefault bool
}

// ConfigStore is the on-disk JSON structure for all profiles.
type ConfigStore struct {
	Profiles       map[string]Config `json:"profiles"`
	DefaultProfile string            `json:"defaultProfile,omitempty"`
}

// ConfigurationError represents a configuration-related error.
type ConfigurationError struct {
	Msg string
}

func (e *ConfigurationError) Error() string { return e.Msg }

// ValidationError represents a validation-related error.
type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

// Option configures a ProfileConfigManager.
type Option func(*ProfileConfigManager)

// WithConfigDir sets the configuration directory.
func WithConfigDir(dir string) Option {
	return func(m *ProfileConfigManager) {
		m.configPath = filepath.Join(dir, configFileName)
	}
}

// WithCryptoService injects a custom TokenCryptoService.
func WithCryptoService(cs *TokenCryptoService) Option {
	return func(m *ProfileConfigManager) {
		m.cryptoService = cs
	}
}

// ProfileConfigManager manages profile-based configuration stored on disk.
type ProfileConfigManager struct {
	configPath    string
	cryptoService *TokenCryptoService
}

// NewProfileConfigManager creates a new manager with the given options.
// Default configDir is ~/.slack-cli.
func NewProfileConfigManager(opts ...Option) *ProfileConfigManager {
	home, _ := os.UserHomeDir()
	m := &ProfileConfigManager{
		configPath:    filepath.Join(home, defaultConfigDirName, configFileName),
		cryptoService: NewTokenCryptoService(),
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

// TokenKind indicates which slot a token was saved to.
type TokenKind string

const (
	TokenKindBot  TokenKind = "Bot"
	TokenKindUser TokenKind = "User"
)

// SetToken encrypts and saves a token for the given profile.
// The token prefix (xoxb- / xoxp-) determines which slot is used.
// Returns the detected TokenKind.
func (m *ProfileConfigManager) SetToken(token, profile string) (TokenKind, error) {
	store, err := m.getConfigStore()
	if err != nil {
		return "", err
	}

	profileName := m.resolveProfileName(profile, store)

	encrypted, err := m.cryptoService.Encrypt(token)
	if err != nil {
		return "", err
	}

	cfg := store.Profiles[profileName]
	cfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	kind := detectTokenKind(token)
	switch kind {
	case TokenKindUser:
		cfg.UserToken = encrypted
	default:
		cfg.BotToken = encrypted
	}

	store.Profiles[profileName] = cfg

	if store.DefaultProfile == "" || profileName == defaultProfileName {
		store.DefaultProfile = profileName
	}

	return kind, m.saveConfigStore(store)
}

func detectTokenKind(token string) TokenKind {
	if strings.HasPrefix(token, "xoxp-") {
		return TokenKindUser
	}
	return TokenKindBot
}

// GetConfig returns the decrypted configuration for the given profile.
// Returns nil, nil if the profile does not exist.
func (m *ProfileConfigManager) GetConfig(profile string) (*Config, error) {
	store, err := m.getConfigStore()
	if err != nil {
		return nil, err
	}

	profileName := m.resolveProfileName(profile, store)

	cfg, ok := store.Profiles[profileName]
	if !ok {
		return nil, nil
	}

	// Copy non-token fields directly; tokens need decryption.
	// Timezone must be included here so that callers (e.g.
	// config get) see the stored value through GetConfig.
	result := Config{
		Timezone:  cfg.Timezone,
		UpdatedAt: cfg.UpdatedAt,
	}

	if cfg.BotToken != "" {
		result.BotToken = m.decryptToken(cfg.BotToken)
	}
	if cfg.UserToken != "" {
		result.UserToken = m.decryptToken(cfg.UserToken)
	}

	// A profile with only a timezone (no tokens) is still valid.
	if result.BotToken == "" && result.UserToken == "" && result.Timezone == "" {
		return nil, nil
	}

	return &result, nil
}

func (m *ProfileConfigManager) resolveProfileName(profile string, store *ConfigStore) string {
	if profile != "" {
		return profile
	}
	if store.DefaultProfile != "" {
		return store.DefaultProfile
	}
	return defaultProfileName
}

// ListProfiles returns all profiles with their configs.
func (m *ProfileConfigManager) ListProfiles() ([]Profile, error) {
	store, err := m.getConfigStore()
	if err != nil {
		return nil, err
	}

	currentProfile := store.DefaultProfile
	if currentProfile == "" {
		currentProfile = defaultProfileName
	}

	profiles := make([]Profile, 0, len(store.Profiles))
	for name, cfg := range store.Profiles {
		profiles = append(profiles, Profile{
			Name:      name,
			Config:    cfg,
			IsDefault: name == currentProfile,
		})
	}
	return profiles, nil
}

// UseProfile switches the default profile. Returns an error if the profile
// does not exist.
func (m *ProfileConfigManager) UseProfile(profile string) error {
	store, err := m.getConfigStore()
	if err != nil {
		return err
	}

	if _, ok := store.Profiles[profile]; !ok {
		return &ConfigurationError{
			Msg: fmt.Sprintf("Profile %q does not exist", profile),
		}
	}

	store.DefaultProfile = profile
	return m.saveConfigStore(store)
}

// GetCurrentProfile returns the name of the current default profile.
func (m *ProfileConfigManager) GetCurrentProfile() (string, error) {
	store, err := m.getConfigStore()
	if err != nil {
		return "", err
	}
	if store.DefaultProfile != "" {
		return store.DefaultProfile, nil
	}
	return defaultProfileName, nil
}

// ClearConfig removes a profile. If the deleted profile was the default,
// a remaining profile is promoted. If no profiles remain the config file
// is deleted.
func (m *ProfileConfigManager) ClearConfig(profile string) error {
	store, err := m.getConfigStore()
	if err != nil {
		return err
	}

	profileName := m.resolveProfileName(profile, store)

	delete(store.Profiles, profileName)

	if store.DefaultProfile == profileName {
		remaining := make([]string, 0, len(store.Profiles))
		for k := range store.Profiles {
			remaining = append(remaining, k)
		}
		if len(remaining) > 0 {
			store.DefaultProfile = remaining[0]
		} else {
			// No profiles left; remove config file.
			if err := os.Remove(m.configPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			return nil
		}
	}

	return m.saveConfigStore(store)
}

// SetTimezone saves the timezone string for the given profile.
func (m *ProfileConfigManager) SetTimezone(
	timezone, profile string,
) error {
	if _, err := time.LoadLocation(timezone); err != nil {
		return &ValidationError{
			Msg: fmt.Sprintf("invalid timezone %q: %v", timezone, err),
		}
	}

	store, err := m.getConfigStore()
	if err != nil {
		return err
	}

	profileName := m.resolveProfileName(profile, store)
	cfg := store.Profiles[profileName]
	cfg.Timezone = timezone
	cfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	store.Profiles[profileName] = cfg

	if store.DefaultProfile == "" || profileName == defaultProfileName {
		store.DefaultProfile = profileName
	}

	return m.saveConfigStore(store)
}

// GetTimezone returns the timezone string for the given profile.
// Returns "" if unset.
func (m *ProfileConfigManager) GetTimezone(profile string) (string, error) {
	store, err := m.getConfigStore()
	if err != nil {
		return "", err
	}

	profileName := m.resolveProfileName(profile, store)
	cfg, ok := store.Profiles[profileName]
	if !ok {
		return "", nil
	}
	return cfg.Timezone, nil
}

// MaskToken masks a token for display, showing only the first and last
// few characters.
func (m *ProfileConfigManager) MaskToken(token string) string {
	if len(token) <= tokenMinLength {
		return "****"
	}
	prefix := token[:tokenMaskLength]
	suffix := token[len(token)-tokenMaskLength:]
	return prefix + "-****-****-" + suffix
}

func (m *ProfileConfigManager) decryptToken(token string) string {
	if m.cryptoService.IsEncrypted(token) {
		decrypted, err := m.cryptoService.Decrypt(token)
		if err != nil {
			return token
		}
		return decrypted
	}
	return token
}

func (m *ProfileConfigManager) getConfigStore() (*ConfigStore, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &ConfigStore{Profiles: make(map[string]Config)}, nil
		}
		return nil, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, &ConfigurationError{Msg: "Invalid config file format"}
	}

	// Check if migration is needed (old format has "token" but no "profiles").
	if m.needsMigration(raw) {
		return m.migrateOldConfig(data)
	}

	var store ConfigStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, &ConfigurationError{Msg: "Invalid config file format"}
	}
	if store.Profiles == nil {
		store.Profiles = make(map[string]Config)
	}
	return &store, nil
}

func (m *ProfileConfigManager) needsMigration(raw map[string]json.RawMessage) bool {
	_, hasToken := raw["token"]
	_, hasProfiles := raw["profiles"]
	return hasToken && !hasProfiles
}

func (m *ProfileConfigManager) migrateOldConfig(data []byte) (*ConfigStore, error) {
	var old struct {
		Token     string `json:"token"`
		UpdatedAt string `json:"updatedAt"`
	}
	if err := json.Unmarshal(data, &old); err != nil {
		return nil, &ConfigurationError{Msg: "Invalid config file format"}
	}

	plainToken := old.Token
	if m.cryptoService.IsEncrypted(old.Token) {
		decrypted, err := m.cryptoService.Decrypt(old.Token)
		if err != nil {
			return nil, err
		}
		plainToken = decrypted
	}

	encrypted, err := m.cryptoService.Encrypt(plainToken)
	if err != nil {
		return nil, err
	}

	cfg := Config{UpdatedAt: old.UpdatedAt}
	if detectTokenKind(plainToken) == TokenKindUser {
		cfg.UserToken = encrypted
	} else {
		cfg.BotToken = encrypted
	}

	newStore := &ConfigStore{
		Profiles:       map[string]Config{defaultProfileName: cfg},
		DefaultProfile: defaultProfileName,
	}

	if err := m.saveConfigStore(newStore); err != nil {
		return nil, err
	}
	return newStore, nil
}

func (m *ProfileConfigManager) saveConfigStore(store *ConfigStore) error {
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, dirPermission); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	payload, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Atomic write: write to temp file, then rename.
	tempPath := fmt.Sprintf("%s.%d.%d.tmp", m.configPath, os.Getpid(), time.Now().UnixNano())
	f, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePermission)
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}

	if _, err := f.Write(payload); err != nil {
		f.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to write config: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to close temp config file: %w", err)
	}

	if err := os.Rename(tempPath, m.configPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp config file: %w", err)
	}
	return nil
}
