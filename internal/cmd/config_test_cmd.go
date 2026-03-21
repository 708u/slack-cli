package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/708u/slack-cli/internal/config"
)

// ConfigTestCmd tests the current tokens by calling auth.test and
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
	tokens, err := config.GetConfigOrError(c.Profile)
	if err != nil {
		return err
	}

	tested := false

	if tokens.BotToken != "" {
		fmt.Println("Bot token:")
		if err := testToken(tokens.BotToken); err != nil {
			fmt.Printf("  Error: %v\n", err)
		}
		tested = true
		fmt.Println()
	}

	if tokens.UserToken != "" {
		fmt.Println("User token:")
		if err := testToken(tokens.UserToken); err != nil {
			fmt.Printf("  Error: %v\n", err)
		}
		tested = true
	}

	if !tested {
		return fmt.Errorf("no tokens configured")
	}

	return nil
}

func testToken(token string) error {
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

	scopes := resp.Header.Get("X-Oauth-Scopes")

	fmt.Printf("  Workspace:   %s (%s)\n", result.Team, result.URL)
	fmt.Printf("  User:        %s\n", result.User)
	if scopes != "" {
		fmt.Println("  Granted scopes:")
		for s := range strings.SplitSeq(scopes, ",") {
			fmt.Printf("    - %s\n", strings.TrimSpace(s))
		}
	}

	return nil
}
