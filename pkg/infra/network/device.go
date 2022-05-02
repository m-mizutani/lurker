package network

import (
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/utils"
)

type Device interface {
	ReadPacket() chan gopacket.Packet
	WritePacket([]byte)
	GetDeviceAddrs() ([]net.Addr, error)
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

func (x *device) WritePacket(pktData []byte) {
	if err := x.handle.WritePacketData(pktData); err != nil {
		utils.HandleError(goerr.Wrap(err, "fail to send ARP reply"))
	}
}

func (x *device) GetDeviceAddrs() ([]net.Addr, error) {
	iface, err := net.InterfaceByName(x.name)
	if err != nil {
		return nil, goerr.Wrap(err, "failed lookup device").With("name", x.name)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, goerr.Wrap(err, "fail to get device address").With("name", x.name)
	}

	return addrs, nil
}
