package network

import (
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/m-mizutani/goerr"
)

type Device interface {
	ReadPacket() chan gopacket.Packet
	WritePacket(gopacket.Packet) error
}

type device struct {
	name   string
	handle *pcap.Handle
	src    *gopacket.PacketSource
}

func New(devName string) (*device, error) {
	var (
		snapshotLen int32 = 0xffff
		promiscuous       = true
		timeout           = time.Microsecond
	)

	handle, err := pcap.OpenLive(devName, snapshotLen, promiscuous, timeout)
	if err != nil {
		return nil, goerr.Wrap(err, "fail to open device")
	}

	src := gopacket.NewPacketSource(handle, handle.LinkType())

	return &device{
		name:   devName,
		handle: handle,
		src:    src,
	}, nil
}

func (x *device) ReadPacket() chan gopacket.Packet {
	return x.src.Packets()
}

func (x *device) WritePacket(pkt gopacket.Packet) error {
	if err := x.handle.WritePacketData(pkt.Data()); err != nil {
		return goerr.Wrap(err, "fail to send ARP reply")
	}

	return nil
}
