package lurker

import (
	"log"
	"time"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type Lurker struct {
	sourceName string

	pcapHandle *pcap.Handle
	
	// handlers
	handlers []Handler
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


func (x *Lurker) AddArpSpoofer() {
	x.handlers = append(x.handlers, NewHandler("arp_spoofer"))
}

func (x *Lurker) AddTcpSpoofer() {
	x.handlers = append(x.handlers, NewHandler("tcp_spoofer"))
}

func (x *Lurker) AddDataLogger() {
	x.handlers = append(x.handlers, NewHandler("data_logger"))
}

func (x *Lurker) AddConnLogger() {
	x.handlers = append(x.handlers, NewHandler("conn_logger"))
}



func (x *Lurker) Loop() error {
	if x.pcapHandle == nil {
		return errors.New("No available device or pcap file, need to specify one of them")
	}
	
	packetSource := gopacket.NewPacketSource(x.pcapHandle, x.pcapHandle.LinkType())
	for packet := range packetSource.Packets() {
		for _, handler := range x.handlers {
			handler.Handle(&packet)
		}
	}

	return nil
}

func (x *Lurker) Close() {
	if x.pcapHandle != nil {
		x.pcapHandle.Close()
	}
}



