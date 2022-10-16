package arp

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/domain/types"
)

type arpHandler struct {
	networks   []*net.IPNet
	deviceAddr net.HardwareAddr
	device     interfaces.Device
}

func New(device interfaces.Device, addr net.HardwareAddr, networks []*net.IPNet) *arpHandler {
	return &arpHandler{
		networks:   networks,
		deviceAddr: addr,
		device:     device,
	}
}

func (x *arpHandler) Handle(ctx *types.Context, pkt gopacket.Packet) error {
	l := extractPktLayers(pkt)
	if l == nil {
		return nil
	}

	if l.arp.Operation != 1 {
		return nil
	}

	if !x.isTarget(l.arp.DstProtAddress) {
		return nil
	}

	reply, err := createARPReply(l, x.deviceAddr)
	if err != nil {
		return err
	}

	x.device.WritePacket(reply)

	return nil
}

func (x *arpHandler) Tick(ctx *types.Context) error {
	return nil
}

func (x *arpHandler) isTarget(ip net.IP) bool {
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

type pktLayers struct {
	eth *layers.Ethernet
	arp *layers.ARP
}

func extractPktLayers(pkt gopacket.Packet) *pktLayers {
	l := &pktLayers{}
	var eth layers.Ethernet
	var arp layers.ARP
	var payload gopacket.Payload

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &arp, &payload)

	var decoded []gopacket.LayerType
	if err := parser.DecodeLayers(pkt.Data(), &decoded); err != nil {
		return nil // ignore
	}

	for _, layerType := range decoded {
		switch layerType {
		case layers.LayerTypeEthernet:
			l.eth = &eth
		case layers.LayerTypeARP:
			l.arp = &arp
		}
	}

	if l.eth == nil || l.arp == nil {
		return nil
	}

	return l
}

func createARPReply(l *pktLayers, deviceAddr net.HardwareAddr) ([]byte, error) {
	var options gopacket.SerializeOptions
	buffer := gopacket.NewSerializeBuffer()

	eth := &layers.Ethernet{
		SrcMAC:       deviceAddr,
		DstMAC:       l.eth.SrcMAC,
		EthernetType: l.eth.EthernetType,
	}

	arp := &layers.ARP{
		AddrType:          l.arp.AddrType,
		Protocol:          l.arp.Protocol,
		HwAddressSize:     l.arp.HwAddressSize,
		ProtAddressSize:   l.arp.ProtAddressSize,
		Operation:         2,
		DstHwAddress:      l.arp.SourceHwAddress,
		SourceHwAddress:   deviceAddr,
		SourceProtAddress: l.arp.DstProtAddress,
		DstProtAddress:    l.arp.SourceProtAddress,
	}

	if err := gopacket.SerializeLayers(buffer, options,
		eth,
		arp,
	); err != nil {
		return nil, goerr.Wrap(err)
	}

	return buffer.Bytes(), nil
}
