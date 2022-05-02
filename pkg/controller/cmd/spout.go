package cmd

import (
	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/infra"
	webhook "github.com/m-mizutani/lurker/pkg/infra/slack"
	"github.com/m-mizutani/lurker/pkg/utils"
	"github.com/slack-go/slack"
)

func configureSpout(cfg *Config, clients *infra.Clients) *interfaces.Spout {
	// spout
	var spoutOptions []interfaces.SpoutOption

	// Slack Webhook
	if cfg.SlackWebhookURL != "" {
		client := webhook.New(cfg.SlackWebhookURL)
		f := func(ctx *types.Context, msg *slack.WebhookMessage) {
			if err := client.Post(ctx, msg); err != nil {
				utils.HandleError(err)
			}
		}
		spoutOptions = append(spoutOptions,
			interfaces.WithSlack(f),
		)
	}

	spout := interfaces.NewSpout(
		clients,
		spoutOptions...,
	)

	return spout
}
