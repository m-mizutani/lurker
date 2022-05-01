package tcp

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/m-mizutani/ttlcache"

	"github.com/m-mizutani/lurker/pkg/service/spout"
)

type tcpHandler struct {
	allowList []net.IPNet
	flows     *ttlcache.CacheTable[uint64, *tcpFlow]
	elapsed   time.Duration
	lastTick  uint64
}

type tcpFlow struct {
	srcHost, dstHost gopacket.Endpoint
	srcPort, dstPort uint16

	baseSeq  uint32
	nextSeq  uint32
	recvData []byte
}

func New(optins ...Option) *tcpHandler {
	hdlr := &tcpHandler{
		flows: ttlcache.New[uint64, *tcpFlow](),
	}

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

	nw := pkt.NetworkLayer()
	if nw == nil {
		return nil
	}
	src, dst := nw.NetworkFlow().Endpoints()

	tp := pkt.TransportLayer()
	hv := flowHash(nw, tp)

	if l.tcp.FIN || l.tcp.RST {
		return nil
	}

	if l.tcp.SYN && !l.tcp.ACK {
		if !x.isInAllowList(net.IP(dst.Raw())) {
			return nil
		}
		flow := &tcpFlow{
			srcHost: src,
			dstHost: dst,
			srcPort: uint16(l.tcp.SrcPort),
			dstPort: uint16(l.tcp.DstPort),

			baseSeq: l.tcp.Seq,
			nextSeq: l.tcp.Seq + 1,
		}
		x.flows.Set(hv, flow, 5)

		synAckPkt, err := createSynAckPacket(l)
		if err != nil {
			return err
		}
		if err := spouts.WritePacket(synAckPkt); err != nil {
			return err
		}

	} else if flow := x.flows.Get(hv); flow != nil {
		if dst.String() == flow.dstHost.String() && l.tcp.DstPort == layers.TCPPort(flow.dstPort) {
			if flow.nextSeq == l.tcp.Seq {
				flow.recvData = append(flow.recvData, l.tcp.Payload...)
				flow.nextSeq += uint32(len(l.tcp.Payload))
			}
		}
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
	x.elapsed += duration
	tick := uint64(x.elapsed / time.Second)
	delta := tick - x.lastTick

	id := x.flows.SetHook(func(flow *tcpFlow) uint64 {
		spouts.Log("expires: %v -> %v:%d")
		return 0
	})
	defer x.flows.DelHook(id)

	x.flows.Elapse(delta)

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
	var payload gopacket.Payload

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &tcp, &payload)

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

func flowHash(nw gopacket.NetworkLayer, tp gopacket.TransportLayer) uint64 {
	nwHash := nw.NetworkFlow().FastHash()
	tpHash := tp.TransportFlow().FastHash()

	b1, b2 := make([]byte, 8), make([]byte, 8)
	binary.LittleEndian.PutUint64(b1, nwHash)
	binary.LittleEndian.PutUint64(b2, tpHash)

	hash := fnv.New64a()
	_, _ = hash.Write(b1)
	_, _ = hash.Write(b2)
	return hash.Sum64()
}
