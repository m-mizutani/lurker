package main

import (
	"time"

	"github.com/google/gopacket"
)

type dataStoreHandler struct {
	flowLogMap map[flowKey]*flowLog
}

type pktLog struct {
	capInfo gopacket.CaptureInfo
	data    []byte
}

type flowLog struct {
	packets []pktLog
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
	if !ok {
		flow = &flowLog{packets: []pktLog{}}
		x.flowLogMap[*fkey] = flow
		x.flowLogMap[fkey.swap()] = flow
	}

	flow.packets = append(flow.packets, log)

	return nil
}

func (x *dataStoreHandler) timer(t time.Time) error {

	return nil
}

func (x *dataStoreHandler) teardown() error {
	return nil
}
