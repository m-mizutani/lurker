package interfaces

import (
	"net"

	"github.com/google/gopacket"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
)

type Emitter interface {
	Emit(ctx *types.Context, flow *model.TCPFlow) error
	Close() error
}

type Emitters []Emitter

func (x Emitters) Emit(ctx *types.Context, flow *model.TCPFlow) (errors []error) {
	for _, emitter := range x {
		if err := emitter.Emit(ctx, flow); err != nil {
			errors = append(errors, err)
		}
	}

	return
}

type Handler interface {
	Handle(ctx *types.Context, pkt gopacket.Packet) error
	Tick(ctx *types.Context) error
}

type Device interface {
	ReadPacket() chan gopacket.Packet
	WritePacket(pktData []byte) error
	GetDeviceAddrs() ([]net.Addr, error)
}
