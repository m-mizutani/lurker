package lurker

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"time"
)

type Lurker struct {
	sourceName string

	pcapHandle   *pcap.Handle
	packetSource *gopacket.PacketSource

	// handlers
	handlers []Handler
	gateway  EmitterGateway
}

//
// Constructor
//
func New() Lurker {
	lurker := Lurker{}
	return lurker
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
		snapshotLen int32         = 0xffff
		promiscuous bool          = true
		timeout     time.Duration = -1 * time.Second
	)

	handle, pcapErr := pcap.OpenLive(devName, snapshotLen, promiscuous, timeout)

	if pcapErr != nil {
		return pcapErr
	}

	x.pcapHandle = handle
	return nil
}

//
// Manage Emitter
//

func (x *Lurker) AddFluentdEmitter(addr string) error {
	emitter, err := NewFluentd(addr, 24224)
	if err != nil {
		return err
	}

	x.AddEmitter(emitter)
	return nil
}

func (x *Lurker) AddEmitter(emitter Emitter) {
	x.gateway.Add(emitter)
}

//
// Manage Handler
//

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

func (x *Lurker) Preprocess() error {
	if x.pcapHandle == nil {
		return errors.New("No available device or pcap file, need to specify one of them")
	}

	if x.packetSource == nil {
		x.packetSource = gopacket.NewPacketSource(x.pcapHandle, x.pcapHandle.LinkType())

		for _, handler := range x.handlers {
			handler.SetEmitterGateway(&x.gateway)
		}
	}

	return nil
}

func (x *Lurker) Loop() error {
	err := x.Preprocess()
	if err != nil {
		return err
	}

	for packet := range x.packetSource.Packets() {
		for _, handler := range x.handlers {
			handler.Handle(&packet)
		}
	}

	return nil
}

func (x *Lurker) Read(readSize int) error {
	err := x.Preprocess()
	if err != nil {
		return err
	}

	count := 0

	for {
		packet, pktErr := x.packetSource.NextPacket()
		if packet == nil {
			return pktErr
		}

		for _, handler := range x.handlers {
			handler.Handle(&packet)
		}

		count += 1

		if count >= readSize {
			return nil
		}
	}
}

func (x *Lurker) Close() {
	if x.pcapHandle != nil {
		x.pcapHandle.Close()
	}
}
