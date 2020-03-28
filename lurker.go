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
	disableArp  bool
	targetAddrs []string

	awsRegion   string
	awsS3Bucket string
}

type packetHandler interface {
	setup() error
	handle(pkt gopacket.Packet) error
	timer(t time.Time) error
	teardown() error
}

type packetHandlers []packetHandler

const (
	pcapSnapshotLen uint32 = 0xffff
)

func (x *packetHandlers) setup() error {
	for _, handler := range *x {
		if err := handler.setup(); err != nil {
			return errors.Wrapf(err, "Fail to setup %s", reflect.TypeOf(handler))
		}
	}

	return nil
}

func (x *packetHandlers) handle(pkt gopacket.Packet) error {
	for _, hdlr := range *x {
		if err := hdlr.handle(pkt); err != nil {
			logger.WithFields(logrus.Fields{
				"packet":      pkt,
				"error":       err,
				"handlerType": reflect.TypeOf(hdlr),
				"handler":     hdlr,
			}).Error("Fail to handle packet")
		}
	}

	return nil
}

func (x *packetHandlers) timer(t time.Time) error {
	for _, hdlr := range *x {
		if err := hdlr.timer(t); err != nil {
			logger.WithFields(logrus.Fields{
				"time":        t,
				"error":       err,
				"handlerType": reflect.TypeOf(hdlr),
				"handler":     hdlr,
			}).Error("Fail to timer operation")
		}
	}

	return nil
}
func (x *packetHandlers) teardown() error {
	for _, hdlr := range *x {
		if err := hdlr.teardown(); err != nil {
			logger.WithFields(logrus.Fields{
				"error":       err,
				"handlerType": reflect.TypeOf(hdlr),
				"handler":     hdlr,
			}).Error("Fail to teardown")
			return errors.Wrapf(err, "handler teardown error %s", reflect.TypeOf(hdlr))
		}
	}

	return nil
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

func (x *lurker) setS3Bucket(region, s3bucket string) {
	x.awsRegion = region
	x.awsS3Bucket = s3bucket
}

func (x *lurker) loop() error {
	if x.pcapHandle == nil {
		return errors.New("No available device or pcap file, need to specify one of them")
	}

	dh := newDataStoreHanlder()
	dh.setS3Bucket(x.awsRegion, x.awsS3Bucket)
	pktHandlers := packetHandlers{dh}

	if !x.dryRun && x.isOnTheFly {
		if !x.disableArp {
			pktHandlers = append(pktHandlers, newArpHandler(x.pcapHandle, x.sourceName, x.targetAddrs))
		}
		pktHandlers = append(pktHandlers, newTcpHandler(x.pcapHandle, x.targetAddrs))
	}

	if err := pktHandlers.setup(); err != nil {
		return err
	}

	packetSource := gopacket.NewPacketSource(x.pcapHandle, x.pcapHandle.LinkType())
	var timestamp *time.Time
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case pkt := <-packetSource.Packets():
			if pkt == nil {
				logger.Debug("teardown handlers")
				return pktHandlers.teardown()
			}

			if !x.isOnTheFly {
				ts := pkt.Metadata().Timestamp

				if timestamp == nil {
					timestamp = &time.Time{}
					*timestamp = ts
				} else {
					for ; ts.Sub(*timestamp) > time.Second; *timestamp = timestamp.Add(time.Second) {
						if err := pktHandlers.timer(*timestamp); err != nil {
							return errors.Wrap(err, "Fail in timer process")
						}
					}
				}
			}

			if err := pktHandlers.handle(pkt); err != nil {
				return err
			}

		case ts := <-ticker.C:
			if x.isOnTheFly {
				if err := pktHandlers.timer(ts); err != nil {
					return errors.Wrap(err, "Fail in internval timer")
				}
			}
		}
	}
}

func (x *lurker) close() {
	if x.pcapHandle != nil {
		x.pcapHandle.Close()
	}
}
