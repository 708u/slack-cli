package config

import "fmt"

// GetConfigOrError returns the decrypted token for the given profile, or an
// error if no configuration exists. When no ProfileConfigManager is supplied a
// default instance is created.
func GetConfigOrError(profile string, mgr ...*ProfileConfigManager) (string, error) {
	var m *ProfileConfigManager
	if len(mgr) > 0 && mgr[0] != nil {
		m = mgr[0]
	} else {
		m = NewProfileConfigManager()
	}

	cfg, err := m.GetConfig(profile)
	if err != nil {
		return "", err
	}

	if cfg == nil {
		profiles, listErr := m.ListProfiles()
		if listErr != nil {
			return "", listErr
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
		return "", &ConfigurationError{
			Msg: fmt.Sprintf(
				"No configuration found for profile %q. Use \"slack-cli config set --token <token> --profile %s\" to set up.",
				profileName, profileName,
			),
		}
	}

	return cfg.Token, nil
}
