package cmd

import (
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/emitters/bq"
	"github.com/m-mizutani/lurker/pkg/emitters/firehose"
	"github.com/m-mizutani/lurker/pkg/emitters/slack"
	"github.com/m-mizutani/lurker/pkg/handlers/arp"
	"github.com/m-mizutani/lurker/pkg/handlers/tcp"
	"github.com/m-mizutani/lurker/pkg/infra/network"
	"github.com/m-mizutani/lurker/pkg/utils"
	"github.com/m-mizutani/zlog"

	"github.com/urfave/cli/v2"
)

type Config struct {
	NetworkDevice string
	ListenAddrs   cli.StringSlice
	ExcludePorts  cli.IntSlice
	ArpSpoof      bool
}

func Run(argv []string) error {
	var (
		cfg Config

		slackWebhookURL    string
		bigQueryProjectID  string
		bigQueryDataset    string
		firehoseRegion     string
		firehoseStreamName string

		logLevel string
	)

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
			&cli.IntSliceFlag{
				Name:    "exclude-ports",
				Usage:   "Exclude port numbers",
				Aliases: []string{"e"},
				EnvVars: []string{"LURKER_EXCLUDE_PORTS"},
			},
			&cli.BoolFlag{
				Name:        "arp-spoof",
				Usage:       "Enable ARP spoofing",
				Destination: &cfg.ArpSpoof,
				Aliases:     []string{"a"},
				EnvVars:     []string{"LURKER_ARP_SPOOF"},
			},

			// output options
			&cli.StringFlag{
				Name:        "slack-webhook-url",
				Usage:       "Slack incoming webhook URL",
				Destination: &slackWebhookURL,
				EnvVars:     []string{"LURKER_SLACK_WEBHOOK"},
			},
			&cli.StringFlag{
				Name:        "bigquery-project-id",
				Usage:       "BigQuery Project ID",
				Destination: &bigQueryProjectID,
				EnvVars:     []string{"LURKER_BIGQUERY_PROJECT_ID"},
			},
			&cli.StringFlag{
				Name:        "bigquery-dataset",
				Usage:       "BigQuery Dataset name",
				Destination: &bigQueryDataset,
				EnvVars:     []string{"LURKER_BIGQUERY_DATASET"},
			},
			&cli.StringFlag{
				Name:        "firehose-region",
				Usage:       "Amazon Kinesis Data Firehose region",
				Destination: &firehoseRegion,
				EnvVars:     []string{"LURKER_FIREHOSE_REGION"},
			},
			&cli.StringFlag{
				Name:        "firehose-stream-name",
				Usage:       "Amazon Kinesis Data Firehose stream name",
				Destination: &firehoseStreamName,
				EnvVars:     []string{"LURKER_FIREHOSE_STREAM_NAME"},
			},

			// utilities
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Destination: &logLevel,
				EnvVars:     []string{"LURKER_LOG_LEVEL"},
				Value:       "info",
			},
		},
		Before: func(ctx *cli.Context) error {
			utils.InitLogger(zlog.WithLogLevel(logLevel))
			return nil
		},
		Action: func(ctx *cli.Context) error {
			dev, err := network.New(cfg.NetworkDevice)
			if err != nil {
				return goerr.Wrap(err, "failed to configure network device").With("device", cfg.NetworkDevice)
			}

			var tcpOptions []tcp.Option

			targetAddrs, err := configureAddrs(&cfg, dev)
			if err != nil {
				return err
			}
			tcpOptions = append(tcpOptions, addrToTcpOption(targetAddrs)...)

			deviceAddr, err := lookupHWAddr(cfg.NetworkDevice)
			if err != nil {
				return err
			}

			{
				var emitters interfaces.Emitters
				if bigQueryProjectID != "" || bigQueryDataset != "" {
					if bigQueryProjectID == "" || bigQueryDataset == "" {
						return goerr.New("both of bigquery-project-id and bigquery-dataset are required")
					}

					emitter, err := bq.New(bigQueryProjectID, bigQueryDataset)
					if err != nil {
						return err
					}
					emitters = append(emitters, emitter)
				}

				if firehoseRegion != "" || firehoseStreamName != "" {
					if firehoseRegion == "" {
						return goerr.New("firehose-region option required")
					}
					if firehoseStreamName == "" {
						return goerr.New("firehose-stream-name option required")
					}

					emitters = append(emitters, firehose.New(firehoseRegion, firehoseStreamName))
				}

				if slackWebhookURL != "" {
					emitters = append(emitters, slack.New(slackWebhookURL))
				}

				tcpOptions = append(tcpOptions, tcp.WithEmitters(emitters))
			}

			{
				portOptions, err := configureExcludePorts(ctx.IntSlice("exclude-ports"))
				if err != nil {
					return err
				}
				tcpOptions = append(tcpOptions, portOptions...)
			}

			handlers := []interfaces.Handler{
				tcp.New(dev, tcpOptions...),
			}

			if cfg.ArpSpoof {
				handlers = append(handlers, arp.New(dev, deviceAddr, targetAddrs))
			}

			if err := loop(dev, handlers); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(argv); err != nil {
		utils.Logger.Error(err.Error())
		utils.Logger.Err(err).Debug("Error details")
		return err
	}

	return nil
}
