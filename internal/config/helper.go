package config

import "fmt"

// TokenPair holds the resolved bot and user tokens for a profile.
// At least one of BotToken or UserToken is non-empty.
type TokenPair struct {
	BotToken  string
	UserToken string
}

// GetConfigOrError returns the decrypted tokens for the given profile,
// or an error if no configuration exists.
func GetConfigOrError(profile string, mgr ...*ProfileConfigManager) (*TokenPair, error) {
	var m *ProfileConfigManager
	if len(mgr) > 0 && mgr[0] != nil {
		m = mgr[0]
	} else {
		m = NewProfileConfigManager()
	}

	cfg, err := m.GetConfig(profile)
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		profiles, listErr := m.ListProfiles()
		if listErr != nil {
			return nil, listErr
		}
		profileName := profile
		if profileName == "" {
			for _, p := range profiles {
				if p.IsDefault {
					profileName = p.Name
					break
				}
			}
		}
		if profileName == "" {
			profileName = defaultProfileName
		}
		return nil, &ConfigurationError{
			Msg: fmt.Sprintf(
				"No configuration found for profile %q. Use \"slack-cli config set\" to set up.",
				profileName,
			),
		}
	}

	return &TokenPair{
		BotToken:  cfg.BotToken,
		UserToken: cfg.UserToken,
	}, nil
}
