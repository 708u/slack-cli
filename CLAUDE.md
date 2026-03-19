# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code)
when working with code in this repository.

## Build and Test

```bash
make build                                         # out/slack-cli
make test                                          # go test ./...
go test ./internal/slack/ -run TestSendMessage -v  # single test
make vet                                           # go vet ./...
go run ./cmd/slack-cli --help
```

## Architecture

Go CLI for Slack, built with Kong (struct-tag driven) and
slack-go/slack.

```txt
cmd/slack-cli/main.go   # kong.Parse(&cli) entrypoint
internal/
  cmd/                  # Kong command structs, each has Run() error
  config/               # Profile-based token management (AES-256-GCM)
  slack/                # Slack API wrapper
    client.go           # Facade: delegates to *Ops structs
    channel.go          # ChannelOps (+ channel name resolution cache)
    message.go          # MessageOps (send/history/unread/permalink)
    user.go search.go   # UserOps, SearchOps
    reaction.go pin.go star.go reminder.go file.go canvas.go
    types.go            # All domain types
  format/               # Output formatters (table/simple/json)
  util/                 # Pure helpers (sanitize, date, mention, channel)
```

**Key patterns:**

- Every `internal/cmd/*.go` `Run()` calls
  `config.GetConfigOrError(profile)` then `slack.NewClient(token)`.
- `slack.Client` is a thin facade; real logic lives in `*Ops` structs
  (`ChannelOps`, `MessageOps`, etc.) which hold `*slackgo.Client`.
- `ChannelOps.ResolveChannelID` converts name-or-ID to ID, used by
  most operations. Caches channel list on first call.
- `ReminderOps` and `CanvasOps` store a raw `token` + `baseURL` for
  Slack endpoints not covered by slack-go (direct `http.PostForm`).
- `format.Format` enum (`Table`, `Simple`, `JSON`) with `ParseFormat`
  and `PrintJSON` helpers.

## Testing

Tests use `httptest.Server` to mock the Slack API.
`slack.OptionAPIURL(mockURL)` redirects slack-go HTTP calls.
`ReminderOps.baseURL` / `CanvasOps.baseURL` are overridden in tests
for their raw HTTP calls.

Test channel IDs must match `[CDG][A-Z0-9]{8,}` (e.g. `CTEST00001`)
to pass `ResolveChannelID` without triggering a name lookup.

## Adding a New Command

1. Add method to `*Ops` struct in `internal/slack/`
2. Add delegation method to `client.go`
   (missing this causes a runtime error)
3. Create Kong struct with `Run() error` in
   `internal/cmd/<name>.go`
4. Add field to `CLI` struct in `internal/cmd/root.go`
5. Add httptest mock test in `internal/slack/*_test.go`

## Config / Token

```bash
# Set token (interactive prompt, no shell history leak)
slack-cli config set

# Set token via stdin (scripting)
echo "$TOKEN" | slack-cli config set --token-stdin

# Set token for a named profile
slack-cli config set --profile work

# Switch profile
slack-cli config use work

# Environment variable fallback (checked if no --token/--token-stdin)
export SLACK_CLI_TOKEN=xoxb-...

# Master key override (skips key file)
export SLACK_CLI_MASTER_KEY=my-secret
```

Tokens are AES-256-GCM encrypted in `~/.slack-cli/config.json`.
Master key stored in `~/.slack-cli-secrets/master.key`
(auto-generated on first use).
Multi-profile support via `--profile` flag on any command.

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
