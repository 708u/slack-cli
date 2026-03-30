package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/708u/slack-cli/internal/config"
	"github.com/fatih/color"
	"golang.org/x/term"
)

// ConfigCmd groups configuration subcommands.
type ConfigCmd struct {
	Set      ConfigSetCmd      `cmd:"" help:"Set API token"`
	Get      ConfigGetCmd      `cmd:"" help:"Show current configuration"`
	Test     ConfigTestCmd     `cmd:"" help:"Test token and show granted scopes"`
	Profiles ConfigProfilesCmd `cmd:"" help:"List all profiles"`
	Use      ConfigUseCmd      `cmd:"" help:"Switch to a different profile"`
	Current  ConfigCurrentCmd  `cmd:"" help:"Show current active profile"`
	Clear    ConfigClearCmd    `cmd:"" help:"Clear configuration"`
}

// ConfigSetCmd sets the API token for a profile.
type ConfigSetCmd struct {
	Token      string `name:"token" help:"Slack API token (may leak via shell history)" optional:""`
	TokenStdin bool   `name:"token-stdin" help:"Read token from stdin"`
	Timezone   string `name:"timezone" help:"IANA timezone (e.g. Asia/Tokyo)" optional:""`
}

// Run executes the config set command.
func (c *ConfigSetCmd) Run(cli *CLI) error {
	mgr := config.NewProfileConfigManager()
	profileName, err := resolveProfileName(mgr, cli.Profile)
	if err != nil {
		return err
	}

	if c.Timezone != "" {
		if err := mgr.SetTimezone(c.Timezone, cli.Profile); err != nil {
			return err
		}
		fmt.Println(color.GreenString(
			"Timezone %q saved for profile %q", c.Timezone, profileName,
		))
	}

	// Skip token prompt when only --timezone was specified.
	if c.Token == "" && !c.TokenStdin && c.Timezone != "" {
		return nil
	}

	token, err := c.resolveToken()
	if err != nil {
		return err
	}

	kind, err := mgr.SetToken(token, cli.Profile)
	if err != nil {
		return err
	}

	fmt.Println(color.GreenString(
		"%s token saved successfully for profile %q", kind, profileName,
	))
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

	// Read one byte at a time, echo '*' for each character.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback: terminal doesn't support raw mode, read normally.
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return "", fmt.Errorf("token input cancelled")
		}
		token := strings.TrimSpace(scanner.Text())
		if token == "" {
			return "", fmt.Errorf("token cannot be empty")
		}
		return token, nil
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	var buf []byte
	b := make([]byte, 1)
	for {
		if _, err := os.Stdin.Read(b); err != nil {
			return "", fmt.Errorf("failed to read token: %w", err)
		}
		switch b[0] {
		case '\r', '\n':
			fmt.Fprint(os.Stderr, "\r\n")
			token := strings.TrimSpace(string(buf))
			if token == "" {
				return "", fmt.Errorf("token cannot be empty")
			}
			return token, nil
		case 127, '\b': // backspace / delete
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Fprint(os.Stderr, "\b \b")
			}
		case 3: // ctrl-c
			fmt.Fprint(os.Stderr, "\r\n")
			return "", fmt.Errorf("token input cancelled")
		default:
			buf = append(buf, b[0])
			fmt.Fprint(os.Stderr, "*")
		}
	}
}

// resolveProfileName returns the effective profile name for display purposes.
func resolveProfileName(mgr *config.ProfileConfigManager, profile string) (string, error) {
	if profile != "" {
		return profile, nil
	}
	return mgr.GetCurrentProfile()
}

// ConfigGetCmd shows the current configuration for a profile.
type ConfigGetCmd struct{}

// Run executes the config get command.
func (c *ConfigGetCmd) Run(cli *CLI) error {
	mgr := config.NewProfileConfigManager()
	profileName, err := resolveProfileName(mgr, cli.Profile)
	if err != nil {
		return err
	}

	cfg, err := mgr.GetConfig(cli.Profile)
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
	if cfg.BotToken != "" {
		fmt.Printf("  Bot token:  %s\n", color.CyanString(mgr.MaskToken(cfg.BotToken)))
	}
	if cfg.UserToken != "" {
		fmt.Printf("  User token: %s\n", color.CyanString(mgr.MaskToken(cfg.UserToken)))
	}
	if cfg.Timezone != "" {
		fmt.Printf("  Timezone:   %s\n", color.CyanString(cfg.Timezone))
	}
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
		var tokens []string
		if p.Config.BotToken != "" {
			tokens = append(tokens, "bot:"+mgr.MaskToken(p.Config.BotToken))
		}
		if p.Config.UserToken != "" {
			tokens = append(tokens, "user:"+mgr.MaskToken(p.Config.UserToken))
		}
		tokenDisplay := strings.Join(tokens, ", ")
		if tokenDisplay == "" {
			tokenDisplay = "(no tokens)"
		}
		fmt.Printf("  %s %s (%s)\n", marker, color.CyanString(p.Name), tokenDisplay)
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
type ConfigClearCmd struct{}

// Run executes the config clear command.
func (c *ConfigClearCmd) Run(cli *CLI) error {
	mgr := config.NewProfileConfigManager()
	profileName, err := resolveProfileName(mgr, cli.Profile)
	if err != nil {
		return err
	}

	if err := mgr.ClearConfig(cli.Profile); err != nil {
		return err
	}

	fmt.Println(color.GreenString("Profile %q cleared successfully", profileName))
	return nil
}
