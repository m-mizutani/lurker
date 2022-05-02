package interfaces

import (
	"github.com/google/gopacket"
)

type Handler interface {
	Handle(pkt gopacket.Packet, spouts *Spout) error
	Tick(spouts *Spout) error
}
