package tcp

import (
	"math/rand"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/m-mizutani/lurker/pkg/service/spout"
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

func (x *tcpHandler) Handle(pkt gopacket.Packet, spouts *spout.Spouts) error {
	l := extractPktLayers(pkt)
	if l == nil {
		return nil
	}

	if !l.tcp.SYN || l.tcp.ACK {
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

	if err := spouts.Log("Recv SYN: %v:%d -> %v:%d\n", src, l.tcp.SrcPort, dst, l.tcp.DstPort); err != nil {
		return err
	}

	synAckPkt, err := createSynAckPacket(l)
	if err != nil {
		return err
	}
	if err := spouts.WritePacket(synAckPkt); err != nil {
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

func (x *tcpHandler) Elapse(duration time.Duration, spouts *spout.Spouts) error {
	return nil
}

type pktLayers struct {
	eth  *layers.Ethernet
	ipv4 *layers.IPv4
	tcp  *layers.TCP
}

func extractPktLayers(pkt gopacket.Packet) *pktLayers {
	l := &pktLayers{}
	var eth layers.Ethernet
	var ip4 layers.IPv4
	var tcp layers.TCP
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &tcp)

	var decoded []gopacket.LayerType
	if err := parser.DecodeLayers(pkt.Data(), &decoded); err != nil {
		return nil // ignore
	}

	for _, layerType := range decoded {
		switch layerType {
		case layers.LayerTypeEthernet:
			l.eth = &eth
		case layers.LayerTypeIPv4:
			l.ipv4 = &ip4
		case layers.LayerTypeTCP:
			l.tcp = &tcp
		}
	}

	if l.eth == nil || l.ipv4 == nil || l.tcp == nil {
		return nil
	}

	return l
}

func createSynAckPacket(l *pktLayers) ([]byte, error) {
	options := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	buffer := gopacket.NewSerializeBuffer()

	newEtherLayer := &layers.Ethernet{
		SrcMAC:       l.eth.DstMAC,
		DstMAC:       l.eth.SrcMAC,
		EthernetType: l.eth.EthernetType,
	}

	newIpv4Layer := &layers.IPv4{
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
		SrcIP:    l.ipv4.DstIP,
		DstIP:    l.ipv4.SrcIP,
	}
	newTCPLayer := &layers.TCP{
		SrcPort: l.tcp.DstPort,
		DstPort: l.tcp.SrcPort,
		Ack:     l.tcp.Seq + 1,
		Seq:     rand.Uint32(),
		SYN:     true,
		ACK:     true,
		Window:  65535,
	}
	if err := newTCPLayer.SetNetworkLayerForChecksum(newIpv4Layer); err != nil {
		return nil, err
	}

	if err := gopacket.SerializeLayers(buffer, options,
		newEtherLayer,
		newIpv4Layer,
		newTCPLayer,
	); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
