package spout

import (
	"fmt"

	"github.com/m-mizutani/lurker/pkg/infra"
)

type LogFunc func(format string, args ...any) error
type WritePacketFunc func([]byte) error

type Spouts struct {
	Log         LogFunc
	WritePacket WritePacketFunc
}

func New(clients *infra.Clients, options ...SpoutOption) *Spouts {
	output := &Spouts{
		Log: func(format string, args ...any) error {
			fmt.Printf(format, args...)
			return nil
		},
		WritePacket: clients.Device().WritePacket,
	}

	for _, opt := range options {
		opt(output)
	}

	return output
}

type SpoutOption func(*Spouts)

func WithLog(f LogFunc) SpoutOption {
	return func(s *Spouts) {
		s.Log = f
	}
}

func WithWritePacket(f WritePacketFunc) SpoutOption {
	return func(s *Spouts) {
		s.WritePacket = f
	}
}
