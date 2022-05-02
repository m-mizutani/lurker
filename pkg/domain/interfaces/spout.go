package interfaces

import (
	"fmt"
	"time"

	"github.com/m-mizutani/lurker/pkg/infra"
)

type ConsoleFunc func(msg string)
type WritePacketFunc func([]byte)

type Spout struct {
	Console     ConsoleFunc
	WritePacket WritePacketFunc
}

func NewSpout(clients *infra.Clients, options ...SpoutOption) *Spout {
	output := &Spout{
		Console: func(msg string) {
			fmt.Printf("[%s] %s\n", time.Now().Format("2006-01-02T15:04:05.000"), msg)
		},
		WritePacket: clients.Device().WritePacket,
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
