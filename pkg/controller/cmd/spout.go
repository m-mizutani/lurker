package cmd

import (
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/infra"
	"github.com/m-mizutani/lurker/pkg/infra/bq"
	webhook "github.com/m-mizutani/lurker/pkg/infra/slack"
	"github.com/m-mizutani/lurker/pkg/utils"
	"github.com/slack-go/slack"
)

func configureSpout(cfg *Config, clients *infra.Clients) (*interfaces.Spout, error) {
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

	// BigQuery
	if cfg.BigQueryProjectID != "" && cfg.BigQueryDataset != "" {
		client, err := bq.New(cfg.BigQueryProjectID, cfg.BigQueryDataset)
		if err != nil {
			return nil, err
		}
		spoutOptions = append(spoutOptions, interfaces.WithInsertTcpData(
			func(ctx *types.Context, data *model.SchemaTcpData) {
				if err := client.InsertTcpData(ctx, data); err != nil {
					utils.HandleError(err)
				}
			},
		))

	} else if cfg.BigQueryProjectID != "" || cfg.BigQueryDataset != "" {
		return nil, goerr.New("both of bigquery-project-id and bigquery-dataset are required")
	}

	spout := interfaces.NewSpout(
		clients,
		spoutOptions...,
	)

	return spout, nil
}