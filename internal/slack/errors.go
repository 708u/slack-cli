package slack

import (
	"errors"
	"fmt"

	slackgo "github.com/slack-go/slack"
)

// wrapSlackError enriches Slack API errors with actionable hints.
// If the underlying error is "not_allowed_token_type", a message
// explaining that a User token (xoxp-) is required is returned.
func wrapSlackError(op string, err error) error {
	if err == nil {
		return nil
	}

	var slackErr slackgo.SlackErrorResponse
	if errors.As(err, &slackErr) && slackErr.Err == "not_allowed_token_type" {
		return fmt.Errorf("%s: this command requires a User token (xoxp-), not a Bot token (xoxb-)", op)
	}

	return fmt.Errorf("%s: %w", op, err)
}

// isChannelNotFoundError checks whether err is a Slack API
// channel_not_found error.
func isChannelNotFoundError(err error) bool {
	var slackErr slackgo.SlackErrorResponse
	return errors.As(err, &slackErr) && slackErr.Err == "channel_not_found"
}
