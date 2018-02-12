package lurker

import (
	"fmt"
	"log"
	"time"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	// "github.com/m-mizutani/lurker/lurker"
)

type Lurker struct {
	sourceName string

	pcapHandle *pcap.Handle
	
	// handlers
	arpReply *ArpSpoofer
	tcpAck   *TcpSpoofer
	tcpData  *TcpDataLogger
	tcpConn  *TcpConnLogger
}

func (x *Lurker) SetPcapFile(fileName string) error {
	if x.pcapHandle != nil {
		return errors.New("Already set pcap handler, do not specify multiple capture soruce")
	}
	
	log.Println("read from ", fileName)
	handle, pcapErr := pcap.OpenOffline(fileName)

	if pcapErr != nil {
		return pcapErr
	}
	
	x.pcapHandle = handle
	return nil
}

func (x *Lurker) SetPcapDev(devName string) error {
	if x.pcapHandle != nil {
		return errors.New("Already set pcap handler, do not specify multiple capture soruce")
	}
	
	log.Println("capture from ", devName)

	var (
		snapshotLen int32  = 0xffff
		promiscuous bool   = true
		timeout     time.Duration = -1 * time.Second
	)

	handle, pcapErr := pcap.OpenLive(devName, snapshotLen, promiscuous, timeout)

	if pcapErr != nil {
		return pcapErr
	}
	
	x.pcapHandle = handle
	return nil
}

func (x *Lurker) AddFluentdEmitter(addr string) error {
	return nil
}

func (x *Lurker) AddQueueEmitter() error {
	return nil
}


func (x *Lurker) Loop() error {
	if x.pcapHandle == nil {
		return errors.New("No available device or pcap file, need to specify one of them")
	}
	
	packetSource := gopacket.NewPacketSource(x.pcapHandle, x.pcapHandle.LinkType())
	for packet := range packetSource.Packets() {
		if x.arpReply != nil {
			x.arpReply.Handle(&packet)
		}

		if x.tcpAck != nil {
			x.tcpAck.Handle(&packet)
		}

		if x.tcpConn != nil {
			x.tcpConn.Handle(&packet)
		}	
		
		if x.tcpData != nil {
			x.tcpData.Handle(&packet)
		}
	}

	return nil
}

func (x *Lurker) Close() {
	if x.pcapHandle != nil {
		x.pcapHandle.Close()
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
