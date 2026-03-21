---
name: slack-cli-guide
description: |
  Reference guide for the slack-cli command-line tool (github.com/708u/slack-cli).
  Use this skill when the user asks how to use slack-cli, wants to know available
  commands or options, asks about sending Slack messages from the terminal, managing
  Slack channels/users/reminders/pins/reactions via CLI, or mentions "slack-cli" in
  any context. Also trigger when the user wants to configure Slack API tokens,
  manage workspace profiles, or automate Slack operations from the command line.
---

# slack-cli Command Reference

A Go CLI tool for interacting with the Slack API. Built with Kong.

All commands accept `--profile <name>` to target a specific workspace.
Commands with list output accept `--format table|simple|json` (default: table).

## Setup

```bash
# Interactive token input (recommended, no shell history leak)
slack-cli config set

# Via stdin
echo "$SLACK_TOKEN" | slack-cli config set --token-stdin

# Named profile
slack-cli config set --profile work
slack-cli config use work

# Environment variable fallback
export SLACK_CLI_TOKEN=xoxb-...
```

### Required Scopes

Bot Token (`xoxb-`) scopes:

- `channels:read` -- `channels`, `channel info`, `members`
- `channels:history` -- `history`, `unread`
- `channels:join` -- `join`
- `channels:manage` -- `set-topic/purpose`, `invite`, `leave`
- `groups:read` / `groups:history` -- private channel ops
- `im:read` / `im:history` -- DM ops
- `mpim:read` / `mpim:history` -- group DM ops
- `chat:write` -- `send`, `edit`, `delete`, `scheduled`
- `chat:write.public` -- `send` to unjoined channels
- `files:read` -- `canvas list`
- `files:write` -- `upload`
- `pins:read` / `pins:write` -- `pin` ops
- `reactions:write` -- `reaction` ops
- `users:read` -- `users list/info/presence`
- `users:read.email` -- `users lookup`

User Token (`xoxp-`) only scopes:

- `search:read` -- `search`
- `stars:read` / `stars:write` -- `bookmark` ops
- `reminders:read` / `reminders:write` -- `reminder` ops

## config

| Subcommand | Description |
|---|---|
| `config set` | Set API token. Options: `--token`, `--token-stdin`, `--profile` |
| `config get` | Show current config. Options: `--profile` |
| `config test` | Test token and show granted scopes. Options: `--profile` |
| `config profiles` | List all profiles |
| `config use <profile>` | Switch default profile |
| `config current` | Show active profile name |
| `config clear` | Remove profile. Options: `--profile` |

## send

Send or schedule a message to a channel or DM.

```bash
slack-cli send -c general -m "hello"
slack-cli send --user alice -m "hey"
slack-cli send --user-id U0123456789 -m "hey"
slack-cli send --email alice@co.com -m "hi"
slack-cli send -c general -f message.txt
slack-cli send -c general -m "later" --at 2025-01-01T09:00:00
slack-cli send -c general -m "soon" --after 30
slack-cli send -c general -m "reply" -t 1234567890.123456
```

| Flag | Short | Required | Description |
|---|---|---|---|
| `--channel` | `-c` | one of c/user/user-id/email | Channel name or ID |
| `--user` | | | Send DM by username |
| `--user-id` | | | Send DM by user ID |
| `--email` | | | Send DM by email |
| `--message` | `-m` | m or f | Message text |
| `--file` | `-f` | | File containing message |
| `--thread` | `-t` | | Thread timestamp |
| `--at` | | | Schedule time (Unix or ISO 8601) |
| `--after` | | | Schedule after N minutes |

## channels

```bash
slack-cli channels
slack-cli channels --type all --include-archived --format json
```

| Flag | Default | Description |
|---|---|---|
| `--type` | `public` | public, private, im, mpim, all |
| `--include-archived` | false | Include archived |
| `--limit` | 100 | Max channels |
| `--format` | table | table, simple, json |

## history

```bash
slack-cli history -c general
slack-cli history --user alice
slack-cli history --user-id U0123456789
slack-cli history -c general -n 50 --since "2025-01-01 00:00:00"
slack-cli history -c general -t 1234567890.123456 --with-link
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--channel` | `-c` | one of c/user/user-id | Channel name or ID |
| `--user` | `-u` | | Show DM history by username |
| `--user-id` | | | Show DM history by user ID |
| `--number` | `-n` | 10 | Number of messages |
| `--since` | | | Date filter (YYYY-MM-DD HH:MM:SS) |
| `--thread` | `-t` | | Thread timestamp |
| `--with-link` | | false | Include permalink URLs |
| `--format` | | table | table, simple, json |

## unread

```bash
slack-cli unread
slack-cli unread -c general --mark-read
slack-cli unread --count-only --format json
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--channel` | `-c` | | Specific channel |
| `--count-only` | | false | Show only counts |
| `--limit` | | 50 | Max channels |
| `--mark-read` | | false | Mark as read |
| `--format` | | table | table, simple, json |

## search

```bash
slack-cli search -q "deploy error" --sort timestamp --sort-dir desc
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--query` | `-q` | *required* | Search query |
| `--sort` | | score | score, timestamp |
| `--sort-dir` | | desc | asc, desc |
| `--number` | `-n` | 20 | Results per page |
| `--page` | | 1 | Page number |
| `--format` | | table | table, simple, json |

## edit

```bash
slack-cli edit -c general --ts 1234567890.123456 -m "corrected text"
```

| Flag | Short | Description |
|---|---|---|
| `--channel` | `-c` | *required* Channel name or ID |
| `--ts` | | *required* Message timestamp |
| `--message` | `-m` | *required* New message text |

## delete

```bash
slack-cli delete -c general --ts 1234567890.123456
```

| Flag | Short | Description |
|---|---|---|
| `--channel` | `-c` | *required* Channel name or ID |
| `--ts` | | *required* Message timestamp |

## upload

```bash
slack-cli upload -c general -f report.pdf --title "Q4 Report"
slack-cli upload --user-id U0123456789 -f report.pdf
slack-cli upload -c general --content "print('hello')" --filetype python
```

| Flag | Short | Description |
|---|---|---|
| `--channel` | `-c` | Channel name or ID (one of c/user/user-id) |
| `--user` | `-u` | Upload to DM by username |
| `--user-id` | | Upload to DM by user ID |
| `--file` | `-f` | File path (one of file/content required) |
| `--content` | | Text content as snippet |
| `--filename` | | Override filename |
| `--title` | | File title |
| `--message` | `-m` | Initial comment |
| `--filetype` | | Snippet type (python, javascript, csv, etc.) |
| `--thread` | `-t` | Thread timestamp |

## reaction

```bash
slack-cli reaction add -c general -t 1234567890.123456 -e thumbsup
slack-cli reaction remove -c general -t 1234567890.123456 -e thumbsup
```

| Subcommand | Required flags |
|---|---|
| `reaction add` | `-c`, `-t`, `-e` (channel, timestamp, emoji) |
| `reaction remove` | `-c`, `-t`, `-e` |

## pin

```bash
slack-cli pin add -c general -t 1234567890.123456
slack-cli pin remove -c general -t 1234567890.123456
slack-cli pin list -c general --format json
```

| Subcommand | Required flags | Optional |
|---|---|---|
| `pin add` | `-c`, `-t` | |
| `pin remove` | `-c`, `-t` | |
| `pin list` | `-c` | `--format` |

## users

```bash
slack-cli users list --limit 50
slack-cli users info --id U0123456789
slack-cli users lookup --email alice@company.com
slack-cli users presence --name @alice
slack-cli users presence --id U0123456789
```

| Subcommand | Required | Optional |
|---|---|---|
| `users list` | | `--limit` (100), `--format` |
| `users info` | `--id` | `--format` |
| `users lookup` | `--email` | `--format` |
| `users presence` | `--id` or `--name` | `--format` |

## channel

```bash
slack-cli channel info -c general
slack-cli channel set-topic -c general --topic "New topic"
slack-cli channel set-purpose -c general --purpose "New purpose"
```

| Subcommand | Required flags | Optional |
|---|---|---|
| `channel info` | `-c` | `--format` |
| `channel set-topic` | `-c`, `--topic` | |
| `channel set-purpose` | `-c`, `--purpose` | |

## members

```bash
slack-cli members -c general --limit 200 --format json
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--channel` | `-c` | *required* | Channel name or ID |
| `--limit` | | 100 | Max members |
| `--format` | | table | table, simple, json |

## send-ephemeral

```bash
slack-cli send-ephemeral -c general -u U0123456789 -m "Only you see this"
```

| Flag | Short | Description |
|---|---|---|
| `--channel` | `-c` | *required* Channel name or ID |
| `--user` | `-u` | *required* User ID |
| `--message` | `-m` | *required* Message text |
| `--thread` | `-t` | Thread timestamp |

## join / leave / invite

```bash
slack-cli join -c general
slack-cli leave -c general
slack-cli invite -c general -u U001,U002 --force
```

| Command | Required flags | Optional |
|---|---|---|
| `join` | `-c` | |
| `leave` | `-c` | |
| `invite` | `-c`, `-u` (comma-separated IDs) | `--force` |

## reminder

```bash
slack-cli reminder add --text "standup" --after 30
slack-cli reminder add --text "meeting" --at "2025-06-01T10:00:00"
slack-cli reminder list --format json
slack-cli reminder delete --id Rm0123456789
slack-cli reminder complete --id Rm0123456789
```

| Subcommand | Required | Optional |
|---|---|---|
| `reminder add` | `--text`, (`--at` or `--after`) | |
| `reminder list` | | `--format` |
| `reminder delete` | `--id` | |
| `reminder complete` | `--id` | |

## bookmark

```bash
slack-cli bookmark add -c C0123456789 --ts 1234567890.123456
slack-cli bookmark list --limit 50 --format json
slack-cli bookmark remove -c C0123456789 --ts 1234567890.123456
```

| Subcommand | Required | Optional |
|---|---|---|
| `bookmark add` | `-c`, `--ts` | |
| `bookmark list` | | `--limit` (100), `--format` |
| `bookmark remove` | `-c`, `--ts` | |

## canvas

```bash
slack-cli canvas read --id F0123456789
slack-cli canvas list -c general --format json
```

| Subcommand | Required | Optional |
|---|---|---|
| `canvas read` | `--id` (`-i`) | `--format` |
| `canvas list` | `-c` | `--format` |
