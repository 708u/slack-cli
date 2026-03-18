package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/708u/slack-cli/internal/config"
)

// ConfigTestCmd tests the current token by calling auth.test and
// displaying the token type, workspace, user, and granted scopes.
type ConfigTestCmd struct {
	Profile string `name:"profile" help:"Profile name" optional:""`
}

type authTestResult struct {
	OK    bool   `json:"ok"`
	URL   string `json:"url"`
	Team  string `json:"team"`
	User  string `json:"user"`
	Error string `json:"error,omitempty"`
}

// Run executes the config test command.
func (c *ConfigTestCmd) Run() error {
	token, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	// Determine token type from prefix.
	tokenType := "unknown"
	switch {
	case strings.HasPrefix(token, "xoxb-"):
		tokenType = "Bot"
	case strings.HasPrefix(token, "xoxp-"):
		tokenType = "User"
	case strings.HasPrefix(token, "xoxe-"):
		tokenType = "Enterprise"
	}

	// Call auth.test with raw HTTP to read response headers for scopes.
	req, err := http.NewRequest("POST", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return fmt.Errorf("auth test: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth test: %w", err)
	}
	defer resp.Body.Close()

	var result authTestResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("auth test: decode: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("auth test failed: %s", result.Error)
	}

	// x-oauth-scopes header contains the granted scopes.
	scopes := resp.Header.Get("X-Oauth-Scopes")

	fmt.Printf("  Token type:  %s\n", tokenType)
	fmt.Printf("  Workspace:   %s (%s)\n", result.Team, result.URL)
	fmt.Printf("  User:        %s\n", result.User)
	fmt.Println()
	if scopes != "" {
		fmt.Println("  Granted scopes:")
		for _, s := range strings.Split(scopes, ",") {
			fmt.Printf("    - %s\n", strings.TrimSpace(s))
		}
	} else {
		fmt.Println("  Granted scopes: (not available)")
	}

	return nil
}
