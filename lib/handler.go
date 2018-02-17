package lurker

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
)

// Handler is an interface of Lurker packet processing
type Handler interface {
	Handle(packet *gopacket.Packet)
	SetEmitterGateway(gw *EmitterGateway)
	emit(tag string, msg map[string]interface{})
}

type HandlerBase struct {
	gateway *EmitterGateway
}

func (x *HandlerBase) SetEmitterGateway(gateway *EmitterGateway) {
	x.gateway = gateway
}

func (x *HandlerBase) emit(tag string, msg map[string]interface{}) {
	x.gateway.Emit(tag, msg)
}

// NewHandler creates an instance of Handler
func NewHandler(handlerType string) Handler {
	switch handlerType {
	case "arp_spoofer":
		return &ArpSpoofer{}
	case "tcp_spoofer":
		return &TcpSpoofer{}
	case "data_logger":
		return &TcpDataLogger{}
	default:
		return nil
	}
}

type ArpSpoofer struct {
	HandlerBase
}

func (h *ArpSpoofer) Handle(packet *gopacket.Packet) {
	arpLayer := (*packet).Layer(layers.LayerTypeARP)
	if arpLayer != nil {
		arpPkt, _ := arpLayer.(*layers.ARP)

		if arpPkt.Operation == 1 {
			dp := arpPkt.DstProtAddress
			sp := arpPkt.SourceProtAddress
			log := make(map[string]interface{})
			log["src_hw"] = net.HardwareAddr(arpPkt.SourceHwAddress).String()
			if len(dp) == 4 {
				log["dst_pr"] = net.IPv4(dp[0], dp[1], dp[2], dp[3]).String()
			}
			if len(sp) == 4 {
				log["src_pr"] = net.IPv4(sp[0], sp[1], sp[2], sp[3]).String()
			}
			h.emit("log.arp_request", log)
			// fmt.Println("TODO: do action for arp reply")
		}
	}
}

type TcpSpoofer struct {
	HandlerBase
}

func (h *TcpSpoofer) Handle(packet *gopacket.Packet) {
	tcpLayer := (*packet).Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcpPkt, _ := tcpLayer.(*layers.TCP)

		if tcpPkt.FIN == false && tcpPkt.SYN == true &&
			tcpPkt.RST == false && tcpPkt.ACK == false {

			ipv4Layer := (*packet).Layer(layers.LayerTypeIPv4)
			ipv4Pkt, _ := ipv4Layer.(*layers.IPv4)

			log := make(map[string]interface{})
			log["src_addr"] = ipv4Pkt.SrcIP.String()
			log["dst_addr"] = ipv4Pkt.DstIP.String()
			h.emit("log.tcp_syn", log)
		}
	}
}

type TcpDataLogger struct {
	HandlerBase
}

func (h *TcpDataLogger) Handle(packet *gopacket.Packet) {
	tcpLayer := (*packet).Layer(layers.LayerTypeTCP)
	appLayer := (*packet).ApplicationLayer()
	if tcpLayer != nil && appLayer != nil {
		data := appLayer.Payload()
		fmt.Println(data)
	}
}
