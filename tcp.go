package main

import (
	"math/rand"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
)

type tcpHandler struct {
	pcapHandle  *pcap.Handle
	targetAddrs []net.IP
	targets     []string
}

func newTcpHandler(pcapHandle *pcap.Handle) *tcpHandler {
	return &tcpHandler{
		pcapHandle: pcapHandle,
	}
}

func (x *tcpHandler) setup() error {
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

	return outPktData
}
