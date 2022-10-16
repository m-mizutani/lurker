package tcp

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/ttlcache"
)

type tcpHandler struct {
	device   interfaces.Device
	emitters interfaces.Emitters

	networks     []*net.IPNet
	excludePorts map[int16]bool

	flows    *ttlcache.CacheTable[uint64, *model.TCPFlow]
	elapsed  time.Duration
	lastTick uint64
}

func New(dev interfaces.Device, options ...Option) *tcpHandler {
	hdlr := &tcpHandler{
		device:       dev,
		excludePorts: make(map[int16]bool),
		flows:        ttlcache.New[uint64, *model.TCPFlow](),
	}

	for _, opt := range options {
		opt(hdlr)
	}

	_ = hdlr.flows.SetHook(func(flow *model.TCPFlow) uint64 {
		ctx := types.NewContext()
		if err := hdlr.emitters.Emit(ctx, flow); err != nil {

		}

		return 0
	})

	return hdlr
}

type Option func(hdlr *tcpHandler)

func WithEmitters(emitters interfaces.Emitters) Option {
	return func(hdlr *tcpHandler) {
		hdlr.emitters = emitters
	}
}

func WithNetwork(allowed *net.IPNet) Option {
	return func(hdlr *tcpHandler) {
		hdlr.networks = append(hdlr.networks, allowed)
	}
}

func WithExcludePorts(ports []int) Option {
	return func(hdlr *tcpHandler) {
		for _, port := range ports {
			hdlr.excludePorts[int16(port)] = true
		}
	}
}

func (x *tcpHandler) Handle(ctx *types.Context, pkt gopacket.Packet) error {
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
		if !x.isTarget(net.IP(dst.Raw())) {
			return nil
		}
		if x.excludePorts[int16(l.tcp.DstPort)] {
			return nil
		}

		flow := &model.TCPFlow{
			CreatedAt: time.Now(),
			SrcHost:   src,
			DstHost:   dst,
			SrcPort:   uint16(l.tcp.SrcPort),
			DstPort:   uint16(l.tcp.DstPort),

			BaseSeq: l.tcp.Seq,
			NextSeq: l.tcp.Seq + 1,
		}
		x.flows.Set(hv, flow, 5)

		synAckPkt, err := createSynAckPacket(l)
		if err != nil {
			return err
		}

		x.device.WritePacket(synAckPkt)

	} else if flow := x.flows.Get(hv); flow != nil {
		if dst.String() == flow.DstHost.String() && l.tcp.DstPort == layers.TCPPort(flow.DstPort) {
			if flow.NextSeq == l.tcp.Seq {
				flow.RecvAck = true
				flow.RecvData = append(flow.RecvData, l.tcp.Payload...)
				flow.NextSeq += uint32(len(l.tcp.Payload))
			}
		}
	}

	return nil
}

func (x *tcpHandler) isTarget(ip net.IP) bool {
	if x.networks == nil {
		return true
	}

	for _, nw := range x.networks {
		if nw.Contains(ip) {
			return true
		}
	}
	return false
}

func (x *tcpHandler) Tick(ctx *types.Context) error {
	x.flows.Elapse(1)
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
