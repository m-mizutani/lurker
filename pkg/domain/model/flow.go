package model

import (
	"time"

	"github.com/google/gopacket"
)

type TCPFlow struct {
	CreatedAt        time.Time
	SrcHost, DstHost gopacket.Endpoint
	SrcPort, DstPort uint16

	RecvAck  bool
	BaseSeq  uint32
	NextSeq  uint32
	RecvData []byte
}
