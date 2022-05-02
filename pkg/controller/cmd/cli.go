package cmd

import (
	"fmt"
	"net"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/handlers/tcp"
	"github.com/m-mizutani/lurker/pkg/infra"
	"github.com/m-mizutani/lurker/pkg/infra/network"
	"github.com/m-mizutani/lurker/pkg/usecase"

	"github.com/urfave/cli/v2"
)

type Config struct {
	NetworkDevice   string
	SlackWebhookURL string
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
			&cli.StringFlag{
				Name:        "slack-webhook-url",
				Usage:       "Slack incoming webhook URL",
				Destination: &cfg.SlackWebhookURL,
				EnvVars:     []string{"LURKER_SLACK_WEBHOOK"},
			},
		},
		Action: func(ctx *cli.Context) error {
			dev, err := network.New(cfg.NetworkDevice)
			if err != nil {
				return goerr.Wrap(err, "failed to configure network device").With("device", cfg.NetworkDevice)
			}

			clients := infra.New(infra.WithNetworkDevice(dev))

			var tcpOptions []tcp.Option
			addrs, err := dev.GetDeviceAddrs()
			if err != nil {
				return err
			}
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					return goerr.Wrap(err)
				}
				if ip.To4() == nil {
					continue
				}

				tcpOptions = append(tcpOptions, tcp.WithAllowedNetwork(net.IPNet{
					IP:   ip,
					Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0xff),
				}))
			}

			spout := configureSpout(&cfg, clients)

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
