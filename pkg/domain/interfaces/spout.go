package interfaces

import (
	"fmt"

	"github.com/m-mizutani/lurker/pkg/infra"
)

type ConsoleFunc func(format string, args ...any)
type WritePacketFunc func([]byte)

type Spout struct {
	Console     ConsoleFunc
	WritePacket WritePacketFunc
}

func NewSpout(clients *infra.Clients, options ...SpoutOption) *Spout {
	output := &Spout{
		Console: func(format string, args ...any) {
			fmt.Println("Output:", fmt.Sprintf(format, args...))
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
