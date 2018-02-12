package lurker

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Handler interface {
	Handle(packet *gopacket.Packet)
}

func NewHandler(handlerType string) Handler {
	switch handlerType {
	case "arp_spoofer": return &ArpSpoofer{}
	case "tcp_spoofer": return &TcpSpoofer{}
	case "data_logger": return &TcpDataLogger{}
	case "conn_logger": return &TcpConnLogger{}
	default: return nil
	}
}


type ArpSpoofer struct {
}

func (h *ArpSpoofer) Handle (packet *gopacket.Packet) {
	arpLayer := (*packet).Layer(layers.LayerTypeARP)
	if arpLayer != nil {
		arpPkt, _ := arpLayer.(*layers.ARP)

		if arpPkt.Operation == 1 {
			fmt.Println("TODO: do action for arp reply")
		}
	}
}


type TcpSpoofer struct {
}

func (h *TcpSpoofer) Handle (packet *gopacket.Packet) {
	tcpLayer := (*packet).Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcpPkt, _ := tcpLayer.(*layers.TCP)

		if (tcpPkt.FIN == false && tcpPkt.SYN == true &&
			tcpPkt.RST == false && tcpPkt.ACK == false) {
			fmt.Println("TODO: do action for TCP syn packet")
		}
	}
}


type TcpDataLogger struct {
}

func (h *TcpDataLogger) Handle (packet *gopacket.Packet) {
	tcpLayer := (*packet).Layer(layers.LayerTypeTCP)
	appLayer := (*packet).ApplicationLayer()
	if tcpLayer != nil && appLayer != nil {
		data := appLayer.Payload()
		fmt.Println(data)
	}	
}


type TcpConnLogger struct {
}

func (h *TcpConnLogger) Handle (packet *gopacket.Packet) {
	tcpLayer := (*packet).Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcpPkt, _ := tcpLayer.(*layers.TCP)

		if (tcpPkt.FIN == false && tcpPkt.SYN == true &&
			tcpPkt.RST == false && tcpPkt.ACK == false) {
			fmt.Println("TODO: do action for TCP syn packet")
		}
	}
}
