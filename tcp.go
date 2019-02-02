package main

import (
	"math/rand"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pkg/errors"
)

type tcpHandler struct {
	pcapHandle *pcap.Handle
}

func newTcpHandler(pcapHandle *pcap.Handle) *tcpHandler {
	return &tcpHandler{
		pcapHandle: pcapHandle,
	}
}

func (x *tcpHandler) handle(pkt gopacket.Packet) error {
	rawData := createTCPReply(pkt)
	if rawData == nil {
		return nil // nothing to do
	}

	if err := x.pcapHandle.WritePacketData(rawData); err != nil {
		return errors.Wrap(err, "Fail to send TCP reply")
	}

	return nil
}

func (x *tcpHandler) setup() error {
	return nil
}

func createTCPReply(pkt gopacket.Packet) []byte {
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
	ipv4Pkt, _ := ipv4Layer.(*layers.IPv4)

	options := gopacket.SerializeOptions{ComputeChecksums: true}
	buffer := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buffer, options,
		&layers.Ethernet{
			SrcMAC: ethPkt.DstMAC,
			DstMAC: ethPkt.SrcMAC,
		},
		&layers.IPv4{
			SrcIP: ipv4Pkt.DstIP,
			DstIP: ipv4Pkt.SrcIP,
		},
		&layers.TCP{
			SrcPort: tcpPkt.DstPort,
			DstPort: tcpPkt.SrcPort,
			Ack:     tcpPkt.Seq + 1,
			Seq:     rand.Uint32(),
			SYN:     true,
			ACK:     true,
		},
	)
	outPktData := buffer.Bytes()

	return outPktData
}
