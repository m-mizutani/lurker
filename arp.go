package main

import (
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type arpHandler struct {
	pcapHandle  *pcap.Handle
	deviceName  string
	deviceAddr  net.HardwareAddr
	targets     []string
	targetAddrs []net.IP
}

func newArpHandler(pcapHandle *pcap.Handle, deviceName string, targets []string) *arpHandler {
	return &arpHandler{
		pcapHandle: pcapHandle,
		deviceName: deviceName,
		targets:    targets,
	}
}

func (x *arpHandler) setup() error {
	logger.WithField("arpHandler", x).Debug("Setting up")

	dev, err := net.InterfaceByName(x.deviceName)
	if err != nil {
		return errors.Wrapf(err, "Fail to lookup device %s", x.deviceName)
	}

	x.deviceAddr = dev.HardwareAddr

	for _, target := range x.targets {
		addr := net.ParseIP(target)
		if addr == nil {
			return errors.New("Parse error of target IP address " + target)
		}
		logger.WithFields(logrus.Fields{
			"target": target,
		}).Info("Add target IP address to ARP handler")
		x.targetAddrs = append(x.targetAddrs, addr)
	}

	return nil
}

func (x *arpHandler) handle(pkt gopacket.Packet) error {
	rawData := createARPReply(pkt, x.deviceAddr, x.targetAddrs)
	if rawData == nil {
		return nil // nothing to do
	}

	if err := x.pcapHandle.WritePacketData(rawData); err != nil {
		return errors.Wrap(err, "Fail to send ARP reply")
	}

	logger.WithField("raw", rawData).Debug("Sent packet")

	return nil
}

func (x *arpHandler) timer(t time.Time) error { return nil }
func (x *arpHandler) teardown() error         { return nil }

func findIPAddr(addrSet []net.IP, target net.IP) bool {
	for _, addr := range addrSet {
		if addr.Equal(target) {
			return true
		}
	}

	return false
}

func createARPReply(pkt gopacket.Packet, deviceAddr net.HardwareAddr, targets []net.IP) []byte {

	arpLayer := (pkt).Layer(layers.LayerTypeARP)
	if arpLayer == nil {
		return nil
	}

	arpPkt, ok := arpLayer.(*layers.ARP)
	if !ok {
		return nil
	}
	if arpPkt.Operation != 1 {
		return nil
	}
	if len(targets) > 0 && !findIPAddr(targets, arpPkt.DstProtAddress) {
		return nil
	}

	ethLayer := (pkt).Layer(layers.LayerTypeEthernet)
	ethPkt, ok := ethLayer.(*layers.Ethernet)
	if !ok {
		return nil
	}

	logger.WithField("recv ARP", arpPkt).Debug("Creating ARP reply")

	var options gopacket.SerializeOptions
	buffer := gopacket.NewSerializeBuffer()

	gopacket.SerializeLayers(buffer, options,
		&layers.Ethernet{
			SrcMAC:       deviceAddr,
			DstMAC:       ethPkt.SrcMAC,
			EthernetType: ethPkt.EthernetType,
		},
		&layers.ARP{
			AddrType:          arpPkt.AddrType,
			Protocol:          arpPkt.Protocol,
			HwAddressSize:     arpPkt.HwAddressSize,
			ProtAddressSize:   arpPkt.ProtAddressSize,
			Operation:         2,
			DstHwAddress:      arpPkt.SourceHwAddress,
			SourceHwAddress:   deviceAddr,
			SourceProtAddress: arpPkt.DstProtAddress,
			DstProtAddress:    arpPkt.SourceProtAddress,
		},
	)
	outPktData := buffer.Bytes()

	return outPktData
}
