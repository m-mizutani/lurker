package cmd

import (
	"fmt"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/handlers/tcp"
	"github.com/m-mizutani/lurker/pkg/infra"
	"github.com/m-mizutani/lurker/pkg/infra/network"
	"github.com/m-mizutani/lurker/pkg/usecase"

	"github.com/urfave/cli/v2"
)

type Config struct {
	NetworkDevice string
	ListenAddrs   cli.StringSlice

	SlackWebhookURL   string
	BigQueryProjectID string
	BigQueryDataset   string
}

func Run(argv []string) error {
	var cfg Config
	app := &cli.App{
		Name:  "lurker",
		Usage: "Silent network security sensor",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "device",
				Aliases:     []string{"i"},
				Usage:       "Monitoring network device name",
				Destination: &cfg.NetworkDevice,
				EnvVars:     []string{"LURKER_DEVICE"},
				Required:    true,
			},
			&cli.StringSliceFlag{
				Name:        "network",
				Usage:       "Listen network (CIDR format), default is device address",
				Aliases:     []string{"n"},
				Destination: &cfg.ListenAddrs,
				EnvVars:     []string{"LURKER_NETWORK"},
			},

			// spout options
			&cli.StringFlag{
				Name:        "slack-webhook-url",
				Usage:       "Slack incoming webhook URL",
				Destination: &cfg.SlackWebhookURL,
				EnvVars:     []string{"LURKER_SLACK_WEBHOOK"},
			},
			&cli.StringFlag{
				Name:        "bigquery-project-id",
				Usage:       "BigQuery Project ID",
				Destination: &cfg.BigQueryProjectID,
				EnvVars:     []string{"LURKER_BIGQUERY_PROJECT_ID"},
			},
			&cli.StringFlag{
				Name:        "bigquery-dataset",
				Usage:       "BigQuery Dataset name",
				Destination: &cfg.BigQueryDataset,
				EnvVars:     []string{"LURKER_BIGQUERY_DATASET"},
			},
		},
		Action: func(ctx *cli.Context) error {
			dev, err := network.New(cfg.NetworkDevice)
			if err != nil {
				return goerr.Wrap(err, "failed to configure network device").With("device", cfg.NetworkDevice)
			}

			clients := infra.New(infra.WithNetworkDevice(dev))

			var tcpOptions []tcp.Option
			addrs, err := configureAddrs(&cfg, dev)
			if err != nil {
				return err
			}
			for _, ipNet := range addrs {
				tcpOptions = append(tcpOptions, tcp.WithAllowedNetwork(ipNet))
			}

			spout, err := configureSpout(&cfg, clients)
			if err != nil {
				return err
			}

			uc := usecase.New(clients,
				usecase.WithHandler(tcp.New(tcpOptions...)),
				usecase.WithSpout(spout),
			)

			if err := uc.Loop(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(argv); err != nil {
		fmt.Printf("Error: %+v\n", err)
		return err
	}

	return nil
}
