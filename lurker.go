package main

import (
	"reflect"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

type lurker struct {
	sourceName string

	pcapHandle  *pcap.Handle
	isOnTheFly  bool
	dryRun      bool
	targetAddrs []string
}

type packetHandler interface {
	setup() error
	handle(pkt gopacket.Packet) error
}

func newLurker() lurker {
	unit := lurker{
		dryRun: false,
	}
	return unit
}

func (x *lurker) setPcapFile(fileName string) error {
	if x.pcapHandle != nil {
		return errors.New("Already set pcap handler, do not specify multiple capture soruce")
	}

	logger.WithField("fileName", fileName).Debug("Open file")
	handle, err := pcap.OpenOffline(fileName)

	if err != nil {
		return errors.Wrap(err, "Fail to open pcap file")
	}

	x.sourceName = fileName
	x.pcapHandle = handle
	x.isOnTheFly = false
	return nil
}

func (x *lurker) setPcapDev(devName string) error {
	if x.pcapHandle != nil {
		return errors.New("Already set pcap handler, do not specify multiple capture soruce")
	}

	logger.WithField("devName", devName).Debug("Open device")

	var (
		snapshotLen int32 = 0xffff
		promiscuous       = true
		timeout           = -1 * time.Second
	)

	handle, err := pcap.OpenLive(devName, snapshotLen, promiscuous, timeout)
	if err != nil {
		return errors.Wrap(err, "Fail to open device")
	}

	x.sourceName = devName
	x.pcapHandle = handle
	x.isOnTheFly = true
	return nil
}

func (x *lurker) loop() error {
	if x.pcapHandle == nil {
		return errors.New("No available device or pcap file, need to specify one of them")
	}

	var pktHandlers []packetHandler

	if !x.dryRun && x.isOnTheFly {
		pktHandlers = append(pktHandlers, newArpHandler(x.pcapHandle, x.sourceName, x.targetAddrs))
		// pktHandlers = append(pktHandlers, newTcpHandler(x.pcapHandle))
	}

	for _, handler := range pktHandlers {
		if err := handler.setup(); err != nil {
			return errors.Wrapf(err, "Fail to setup %s", reflect.TypeOf(handler))
		}
	}

	packetSource := gopacket.NewPacketSource(x.pcapHandle, x.pcapHandle.LinkType())
	for pkt := range packetSource.Packets() {
		for _, hdlr := range pktHandlers {
			if err := hdlr.handle(pkt); err != nil {
				logger.WithFields(logrus.Fields{
					"packet": pkt,
					"error":  err,
				}).Error("Fail to handle packet")
			}
		}
	}

	return nil
}

func (x *lurker) close() {
	if x.pcapHandle != nil {
		x.pcapHandle.Close()
	}
}
