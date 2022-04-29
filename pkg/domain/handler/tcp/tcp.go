package tcp

import (
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/m-mizutani/lurker/pkg/domain/model"
)

type tcpHandler struct {
	allowList []net.IPNet
}

func New(optins ...Option) *tcpHandler {
	hdlr := &tcpHandler{}

	for _, opt := range optins {
		opt(hdlr)
	}

	return hdlr
}

type Option func(hdlr *tcpHandler)

func WithAllowedNetwork(allowed net.IPNet) Option {
	return func(hdlr *tcpHandler) {
		hdlr.allowList = append(hdlr.allowList, allowed)
	}
}

func (x *tcpHandler) Handle(pkt gopacket.Packet, output *model.Output) error {
	tcpLayer := pkt.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return nil
	}
	tcpPkt, ok := tcpLayer.(*layers.TCP)
	if !ok {
		return nil
	}

	if !tcpPkt.SYN || tcpPkt.ACK {
		return nil
	}

	nw := pkt.NetworkLayer()
	if nw == nil {
		return nil
	}
	src, dst := nw.NetworkFlow().Endpoints()

	if !x.isInAllowList(net.IP(dst.Raw())) {
		return nil
	}

	if err := output.Log("Recv SYN: %v:%d -> %v:%d\n", src, tcpPkt.SrcPort, dst, tcpPkt.DstPort); err != nil {
		return err
	}

	return nil
}

func (x *tcpHandler) isInAllowList(ip net.IP) bool {
	if x.allowList == nil {
		return true
	}

	for _, nw := range x.allowList {
		if nw.Contains(ip) {
			return true
		}
	}
	return false
}

func (x *tcpHandler) Elapse(duration time.Duration, output *model.Output) error {
	return nil
}
