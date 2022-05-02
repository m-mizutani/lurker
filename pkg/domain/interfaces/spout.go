package interfaces

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/infra"
	"github.com/slack-go/slack"
)

type ConsoleFunc func(msg string)
type WritePacketFunc func([]byte)
type SavePcapDataFunc func([]gopacket.Packet)
type SlackFunc func(ctx *types.Context, msg *slack.WebhookMessage)

type Spout struct {
	Console      ConsoleFunc
	WritePacket  WritePacketFunc
	SavePcapData SavePcapDataFunc
	Slack        SlackFunc
}

func NewSpout(clients *infra.Clients, options ...SpoutOption) *Spout {
	output := &Spout{
		Console: func(msg string) {
			fmt.Printf("[%s] %s\n", time.Now().Format("2006-01-02T15:04:05.000"), msg)
		},
		WritePacket:  clients.Device().WritePacket,
		SavePcapData: func(p []gopacket.Packet) {},
		Slack:        func(ctx *types.Context, msg *slack.WebhookMessage) {},
	}

	for _, opt := range options {
		opt(output)
	}

	return output
}

type SpoutOption func(*Spout)

func WithConsole(f ConsoleFunc) SpoutOption {
	return func(s *Spout) {
		s.Console = f
	}
}

func WithWritePacket(f WritePacketFunc) SpoutOption {
	return func(s *Spout) {
		s.WritePacket = f
	}
}

func WithSavePcapData(f SavePcapDataFunc) SpoutOption {
	return func(s *Spout) {
		s.SavePcapData = f
	}
}

func WithSlack(f SlackFunc) SpoutOption {
	return func(s *Spout) {
		s.Slack = f
	}
}
