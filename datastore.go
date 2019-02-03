package main

import (
	"bytes"
	"time"

	"github.com/google/gopacket/layers"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcapgo"
)

type dataStoreHandler struct {
	flowLogMap map[flowKey]*flowLog
	table      timerTable
}

type pktLog struct {
	capInfo gopacket.CaptureInfo
	data    []byte
}

type flowLog struct {
	packets []pktLog
	last    tick
}

func (x flowLog) dumpPcapData() (*bytes.Buffer, error) {
	pcapData := new(bytes.Buffer)

	w := pcapgo.NewWriter(pcapData)
	w.WriteFileHeader(pcapSnapshotLen, layers.LinkTypeEthernet)
	for _, log := range x.packets {
		if err := w.WritePacket(log.capInfo, log.data); err != nil {
			return nil, err
		}
	}

	return pcapData, nil
}

type flowKey struct {
	networkFlow, transportFlow gopacket.Flow
}

func (x flowKey) swap() flowKey {
	newKey := flowKey{
		networkFlow:   x.networkFlow.Reverse(),
		transportFlow: x.transportFlow.Reverse(),
	}
	return newKey
}

func newDataStoreHanlder() *dataStoreHandler {
	ds := dataStoreHandler{
		flowLogMap: make(map[flowKey]*flowLog),
		table:      newTimerTable(180),
	}
	return &ds
}

func getFlowKey(pkt gopacket.Packet) *flowKey {
	netLayer := pkt.NetworkLayer()
	tpLayer := pkt.TransportLayer()

	if netLayer == nil || tpLayer == nil {
		return nil
	}

	return &flowKey{
		networkFlow:   netLayer.NetworkFlow(),
		transportFlow: tpLayer.TransportFlow(),
	}
}

func (x *dataStoreHandler) setup() error {
	return nil
}

func (x *dataStoreHandler) handle(pkt gopacket.Packet) error {
	fkey := getFlowKey(pkt)
	if fkey == nil { // ifpacket is not hashable
		return nil
	}

	log := pktLog{
		capInfo: pkt.Metadata().CaptureInfo,
		data:    pkt.Data(),
	}

	flow, ok := x.flowLogMap[*fkey]
	var initWaitTime tick = 60
	var extendWaitTime tick = 30

	if !ok {
		flow = &flowLog{packets: []pktLog{}, last: x.table.current}
		x.flowLogMap[*fkey] = flow
		x.flowLogMap[fkey.swap()] = flow

		x.table.add(initWaitTime, func(current tick) tick {
			flow, ok := x.flowLogMap[*fkey]
			if !ok {
				logger.WithField("flowKey", fkey).Warn("Missing flow data in map")
				return 0
			}

			// If current is 0, this callback is invoked by flush()
			if current > 0 && current-flow.last < extendWaitTime {
				logger.WithField("key", fkey).Trace("Extended")
				return extendWaitTime // extend
			}

			logger.WithField("key", fkey).Trace("Expired")
			delete(x.flowLogMap, *fkey)
			delete(x.flowLogMap, fkey.swap())

			buf, err := flow.dumpPcapData()
			if err != nil {
				logger.WithError(err).Error("Fail to dump pcap data")
				return 0
			}
			logger.WithField("buf_len", len(buf.Bytes())).Trace("Dump pcap data")

			return 0
		})
	}

	flow.last = x.table.current
	flow.packets = append(flow.packets, log)

	return nil
}

func (x *dataStoreHandler) timer(t time.Time) error {
	x.table.update(1)
	return nil
}

func (x *dataStoreHandler) teardown() error {
	return nil
}
