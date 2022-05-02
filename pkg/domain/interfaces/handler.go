package interfaces

import (
	"github.com/google/gopacket"
	"github.com/m-mizutani/lurker/pkg/domain/types"
)

type Handler interface {
	Handle(ctx *types.Context, pkt gopacket.Packet, spouts *Spout) error
	Tick(ctx *types.Context, spouts *Spout) error
}
