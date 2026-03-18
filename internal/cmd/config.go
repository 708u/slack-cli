package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/708u/slack-cli/internal/config"
	"github.com/fatih/color"
)

// ConfigCmd groups configuration subcommands.
type ConfigCmd struct {
	Set      ConfigSetCmd      `cmd:"" help:"Set API token"`
	Get      ConfigGetCmd      `cmd:"" help:"Show current configuration"`
	Profiles ConfigProfilesCmd `cmd:"" help:"List all profiles"`
	Use      ConfigUseCmd      `cmd:"" help:"Switch to a different profile"`
	Current  ConfigCurrentCmd  `cmd:"" help:"Show current active profile"`
	Clear    ConfigClearCmd    `cmd:"" help:"Clear configuration"`
}

// ConfigSetCmd sets the API token for a profile.
type ConfigSetCmd struct {
	Token      string `name:"token" help:"Slack API token (may leak via shell history)" optional:""`
	TokenStdin bool   `name:"token-stdin" help:"Read token from stdin"`
	Profile    string `name:"profile" help:"Profile name" optional:""`
}

// Run executes the config set command.
func (c *ConfigSetCmd) Run() error {
	token, err := c.resolveToken()
	if err != nil {
		return err
	}

	mgr := config.NewProfileConfigManager()
	profileName, err := resolveProfileName(mgr, c.Profile)
	if err != nil {
		return err
	}

	if err := mgr.SetToken(token, c.Profile); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Token saved successfully for profile %q", profileName))
	return nil
}

func (c *ConfigSetCmd) resolveToken() (string, error) {
	if c.Token != "" && c.TokenStdin {
		return "", fmt.Errorf("cannot use --token and --token-stdin together")
	}

	if c.TokenStdin {
		return readTokenFromStdin(os.Stdin)
	}

	if c.Token != "" {
		fmt.Fprintln(os.Stderr, color.YellowString(
			"Warning: --token may leak secrets via shell history/process list. Prefer --token-stdin or interactive input.",
		))
		return strings.TrimSpace(c.Token), nil
	}

	if envToken := strings.TrimSpace(os.Getenv("SLACK_CLI_TOKEN")); envToken != "" {
		return envToken, nil
	}

	return promptTokenInteractively()
}

func readTokenFromStdin(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read token from stdin: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("no token received from stdin")
	}
	return token, nil
}

func promptTokenInteractively() (string, error) {
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return "", fmt.Errorf(
			"no token provided. Use --token-stdin, set SLACK_CLI_TOKEN, or run this command in an interactive terminal",
		)
	}

	fmt.Fprint(os.Stderr, "Slack API token: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read token: %w", err)
		}
		return "", fmt.Errorf("token input cancelled")
	}

	token := strings.TrimSpace(scanner.Text())
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}
	return token, nil
}

// resolveProfileName returns the effective profile name for display purposes.
func resolveProfileName(mgr *config.ProfileConfigManager, profile string) (string, error) {
	if profile != "" {
		return profile, nil
	}
	return mgr.GetCurrentProfile()
}

// ConfigGetCmd shows the current configuration for a profile.
type ConfigGetCmd struct {
	Profile string `name:"profile" help:"Profile name" optional:""`
}

// Run executes the config get command.
func (c *ConfigGetCmd) Run() error {
	mgr := config.NewProfileConfigManager()
	profileName, err := resolveProfileName(mgr, c.Profile)
	if err != nil {
		return err
	}

	cfg, err := mgr.GetConfig(c.Profile)
	if err != nil {
		return err
	}

	if cfg == nil {
		fmt.Println(color.YellowString(
			"No configuration found for profile %q. Use \"slack-cli config set --token <token> --profile %s\" to set up.",
			profileName, profileName,
		))
		return nil
	}

	bold := color.New(color.Bold)
	bold.Printf("Configuration for profile %q:\n", profileName)
	fmt.Printf("  Token: %s\n", color.CyanString(mgr.MaskToken(cfg.Token)))
	fmt.Printf("  Updated: %s\n", cfg.UpdatedAt)
	return nil
}

// ConfigProfilesCmd lists all profiles.
type ConfigProfilesCmd struct{}

// Run executes the config profiles command.
func (c *ConfigProfilesCmd) Run() error {
	mgr := config.NewProfileConfigManager()
	profiles, err := mgr.ListProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		fmt.Println(color.YellowString(
			"No profiles found. Use \"slack-cli config set --token <token>\" to create one.",
		))
		return nil
	}

	currentProfile, err := mgr.GetCurrentProfile()
	if err != nil {
		return err
	}

	bold := color.New(color.Bold)
	bold.Println("Available profiles:")
	for _, p := range profiles {
		marker := " "
		if p.Name == currentProfile {
			marker = "*"
		}
		maskedToken := mgr.MaskToken(p.Config.Token)
		fmt.Printf("  %s %s (%s)\n", marker, color.CyanString(p.Name), maskedToken)
	}
	return nil
}

// ConfigUseCmd switches to a different profile.
type ConfigUseCmd struct {
	Profile string `arg:"" help:"Profile name to switch to"`
}

// Run executes the config use command.
func (c *ConfigUseCmd) Run() error {
	mgr := config.NewProfileConfigManager()
	if err := mgr.UseProfile(c.Profile); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Switched to profile %q", c.Profile))
	return nil
}

// ConfigCurrentCmd shows the current active profile.
type ConfigCurrentCmd struct{}

// Run executes the config current command.
func (c *ConfigCurrentCmd) Run() error {
	mgr := config.NewProfileConfigManager()
	currentProfile, err := mgr.GetCurrentProfile()
	if err != nil {
		return err
	}

	bold := color.New(color.Bold)
	bold.Printf("Current profile: %s\n", color.CyanString(currentProfile))
	return nil
}

// ConfigClearCmd clears configuration for a profile.
type ConfigClearCmd struct {
	Profile string `name:"profile" help:"Profile name" optional:""`
}

// Run executes the config clear command.
func (c *ConfigClearCmd) Run() error {
	mgr := config.NewProfileConfigManager()
	profileName, err := resolveProfileName(mgr, c.Profile)
	if err != nil {
		return err
	}

	if err := mgr.ClearConfig(c.Profile); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Profile %q cleared successfully", profileName))
	return nil
}
