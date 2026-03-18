# slack-cli

Go CLI tool for interacting with the Slack API.

## Install

```bash
go install github.com/708u/slack-cli@latest
```

Or build from source:

```bash
git clone https://github.com/708u/slack-cli.git
cd slack-cli
go build -o slack-cli .
```

## Setup

Get a Slack Bot Token (`xoxb-...`) from
[api.slack.com/apps](https://api.slack.com/apps) with the scopes
your commands need (e.g. `chat:write`, `channels:read`,
`channels:history`, `users:read`, `search:read`, `pins:write`,
`reactions:write`, `files:write`, `reminders:write`).

```bash
# Interactive input (recommended)
slack-cli config set

# From stdin (CI/scripts)
echo "$SLACK_TOKEN" | slack-cli config set --token-stdin

# Multiple workspaces
slack-cli config set --profile work
slack-cli config use work
```

Tokens are AES-256-GCM encrypted and stored in
`~/.slack-cli/config.json`. The master key is auto-generated at
`~/.slack-cli-secrets/master.key`.

## Usage

```bash
# Send a message
slack-cli send -c general -m "hello"

# Send a DM
slack-cli send --user alice -m "hey"

# Schedule a message
slack-cli send -c general -m "reminder" --after 30

# List channels
slack-cli channels --type all --format json

# Message history
slack-cli history -c general -n 20

# Search
slack-cli search -q "deploy error"

# Unread messages
slack-cli unread --mark-read

# Upload a file
slack-cli upload -c general -f report.pdf

# Reactions
slack-cli reaction add -c general -t 1234567890.123456 -e thumbsup

# Pins
slack-cli pin add -c general -t 1234567890.123456

# Reminders
slack-cli reminder add --text "standup" --after 15

# User info
slack-cli users info --id U0123456789
```

Run `slack-cli --help` for full command list, or
`slack-cli <command> --help` for options.

## Output Formats

Most list commands support `--format table|simple|json`:

- `table` (default) -- human-readable table
- `simple` -- tab-separated, suitable for piping
- `json` -- structured JSON

## Claude Code Plugin

This repository is also a
[Claude Code plugin](https://docs.anthropic.com/en/docs/claude-code/plugins).
It provides a `slack-cli-guide` skill that gives Claude full
knowledge of all commands and options.

```bash
# Add this repo as a marketplace source
/plugin marketplace add 708u/slack-cli

# Install the plugin
/plugin install slack-cli
```

After installation, Claude Code can reference slack-cli commands
when you mention Slack operations in conversation.

## License

MIT
