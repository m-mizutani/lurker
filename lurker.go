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


func HandlePacket(packet gopacket.Packet) {
	// fmt.Println(packet)

	arpLayer := packet.Layer(layers.LayerTypeARP)
	if arpLayer != nil {
		arpPkt, _ := arpLayer.(*layers.ARP)

		if arpPkt.Operation == 1 {
			fmt.Println("TODO: do action for arp reply")
		}
	}

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcpPkt, _ := tcpLayer.(*layers.TCP)

		if (tcpPkt.FIN == false && tcpPkt.SYN == true &&
			tcpPkt.RST == false && tcpPkt.ACK == false) {
			fmt.Println("TODO: do action for TCP syn packet")
		}
	}

	appLayer := packet.ApplicationLayer()
	if tcpLayer != nil && appLayer != nil {
		data := appLayer.Payload()
		fmt.Println(data)
	}
	
	return	
}


func main() {
	var opts struct {
		FileName string `short:"r" description:"A pcap file" value-name:"FILE"`
		DevName string `short:"i" description:"Interface name" value-name:"DEV"`
		FluentDst string `short:"f" description:"Destination of fluentd logs" value-name:"HOST:PORT"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		os.Exit(1)
	}

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

	if pcapErr != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	
	if handle == nil {
		log.Fatal("No available device and pcap file, -i or -r option is mandatory")
		os.Exit(1)
	}

	defer handle.Close()
	
	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		HandlePacket(packet)
	}
}
