package handler

import (
	"time"

	"github.com/google/gopacket"
	"github.com/m-mizutani/lurker/pkg/domain/model"
)

type Handler interface {
	Handle(pkt gopacket.Packet, output *model.Output) error
	Elapse(duration time.Duration, output *model.Output) error
}
