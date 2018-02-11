package main

import (
	"os"
	"fmt"
	"log"
	"time"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/jessevdk/go-flags"
)


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


type PacketHandler struct {
	DevName string

	// handlers
	ArpReply *ArpSpoofer
	TcpAck   *TcpSpoofer
	TcpData  *TcpDataLogger
	TcpConn  *TcpConnLogger
}

func (hdlr *PacketHandler) Read (packet *gopacket.Packet) {
	// fmt.Println(packet)

	if hdlr.ArpReply != nil {
		hdlr.ArpReply.Handle(packet)
	}

	if hdlr.TcpAck != nil {
		hdlr.TcpAck.Handle(packet)
	}

	if hdlr.TcpConn != nil {
		hdlr.TcpConn.Handle(packet)
	}	
	
	if hdlr.TcpData != nil {
		hdlr.TcpData.Handle(packet)
	}
}


type Options struct {
	FileName string `short:"r" description:"A pcap file" value-name:"FILE"`
	DevName string `short:"i" description:"Interface name" value-name:"DEV"`
	FluentDst string `short:"f" description:"Destination of fluentd logs" value-name:"HOST:PORT"`
}

func SetupPcapHandler(opts Options) (*pcap.Handle, error) {
	var handle *pcap.Handle
	var pcapErr error
	
	if opts.FileName != "" {
		log.Println("read from ", opts.FileName)
		fmt.Println(opts.FileName)
		
		handle, pcapErr = pcap.OpenOffline(opts.FileName)
	}
	
	if opts.DevName != "" {
		log.Println("capture from ", opts.DevName)

		var (
			snapshotLen int32  = 0xffff
			promiscuous bool   = true
			timeout     time.Duration = -1 * time.Second
		)

		handle, pcapErr = pcap.OpenLive(opts.DevName, snapshotLen, promiscuous, timeout)
	}

	return handle, pcapErr
}


func main() {
	var opts Options
	
	_, ParseErr := flags.ParseArgs(&opts, os.Args)
	if ParseErr != nil {
		os.Exit(1)
	}

	
	handler := PacketHandler{}
	
	pcapHandle, pcapErr := SetupPcapHandler(opts)
	if pcapErr != nil {
		log.Fatal(pcapErr)
		os.Exit(1)
	}
	
	if pcapHandle == nil {
		log.Fatal("No available device and pcap file, -i or -r option is mandatory")
		os.Exit(1)
	}

	defer pcapHandle.Close()


	handler.ArpReply = &ArpSpoofer{}
	
	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(pcapHandle, pcapHandle.LinkType())
	for packet := range packetSource.Packets() {
		handler.Read(&packet)
		// HandlePacket(packet)
	}
}
