package handler

import (
	"time"

	"github.com/google/gopacket"
	"github.com/m-mizutani/lurker/pkg/service/spout"
)

type Handler interface {
	Handle(pkt gopacket.Packet, spouts *spout.Spouts) error
	Elapse(duration time.Duration, spouts *spout.Spouts) error
}
