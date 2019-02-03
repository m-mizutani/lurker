package main

import (
	"math/rand"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type tcpHandler struct {
	pcapHandle  *pcap.Handle
	targetAddrs []net.IP
	targets     []string
}

func newTcpHandler(pcapHandle *pcap.Handle, targets []string) *tcpHandler {
	return &tcpHandler{
		pcapHandle: pcapHandle,
		targets:    targets,
	}
}

func (x *tcpHandler) setup() error {
	for _, target := range x.targets {
		addr := net.ParseIP(target)
		if addr == nil {
			return errors.New("Parse error of target IP address " + target)
		}
		logger.WithFields(logrus.Fields{
			"target": target,
		}).Info("Add target IP address to TCP handler")
		x.targetAddrs = append(x.targetAddrs, addr)
	}

	return nil
}

func (x *tcpHandler) handle(pkt gopacket.Packet) error {
	rawData := createTCPReply(pkt, x.targetAddrs)
	if rawData == nil {
		return nil // nothing to do
	}

	if err := x.pcapHandle.WritePacketData(rawData); err != nil {
		return errors.Wrap(err, "Fail to send TCP reply")
	}

	return nil
}

func (x *tcpHandler) timer(t time.Time) error { return nil }
func (x *tcpHandler) teardown() error         { return nil }

func createTCPReply(pkt gopacket.Packet, targets []net.IP) []byte {
	tcpLayer := (pkt).Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return nil
	}

	tcpPkt, ok := tcpLayer.(*layers.TCP)
	if !ok {
		return nil
	}

	if tcpPkt.FIN == true || tcpPkt.SYN == false ||
		tcpPkt.RST == true || tcpPkt.ACK == true {
		return nil
	}

	ethLayer := (pkt).Layer(layers.LayerTypeEthernet)
	ethPkt, ok := ethLayer.(*layers.Ethernet)
	if !ok {
		return nil
	}

	ipv4Layer := (pkt).Layer(layers.LayerTypeIPv4)
	ipv4Pkt, ok := ipv4Layer.(*layers.IPv4)
	if !ok {
		return nil
	}

	if len(targets) > 0 && !findIPAddr(targets, ipv4Pkt.DstIP) {
		return nil
	}

	options := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	buffer := gopacket.NewSerializeBuffer()

	newEtherLayer := &layers.Ethernet{
		SrcMAC:       ethPkt.DstMAC,
		DstMAC:       ethPkt.SrcMAC,
		EthernetType: ethPkt.EthernetType,
	}

	newIpv4Layer := &layers.IPv4{
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
		SrcIP:    ipv4Pkt.DstIP,
		DstIP:    ipv4Pkt.SrcIP,
	}
	newTCPLayer := &layers.TCP{
		SrcPort: tcpPkt.DstPort,
		DstPort: tcpPkt.SrcPort,
		Ack:     tcpPkt.Seq + 1,
		Seq:     rand.Uint32(),
		SYN:     true,
		ACK:     true,
		Window:  65535,
	}
	newTCPLayer.SetNetworkLayerForChecksum(newIpv4Layer)

	err := gopacket.SerializeLayers(buffer, options,
		newEtherLayer,
		newIpv4Layer,
		newTCPLayer,
	)
	if err != nil {
		logger.WithError(err).Fatal("fail to create data")
	}

	outPktData := buffer.Bytes()

	logger.WithFields(logrus.Fields{
		"src_addr": ipv4Pkt.SrcIP,
		"dst_addr": ipv4Pkt.DstIP,
		"src_port": tcpPkt.SrcPort,
		"dst_port": tcpPkt.DstPort,
	}).Debug("Created TCP syn+ack reply")

	return outPktData
}
