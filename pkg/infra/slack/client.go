package slack

import (
	"encoding/json"

	"github.com/m-mizutani/lurker/pkg/domain/types"

	"github.com/m-mizutani/goerr"
	"github.com/slack-go/slack"
)

type Client struct {
	webhookURL string
}

func New(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
	}
}

func (x *Client) Post(ctx *types.Context, msg *slack.WebhookMessage) error {
	if err := slack.PostWebhookContext(ctx, x.webhookURL, msg); err != nil {
		raw, _ := json.Marshal(msg)
		return goerr.Wrap(err).With("body", string(raw))
	}

	return nil
}
