package main

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcapgo"
)

type dataStoreHandler struct {
	flowLogMap  map[flowKey]*flowLog
	table       timerTable
	awsRegion   string
	awsS3Bucket string
	wg          sync.WaitGroup
}

type pktLog struct {
	capInfo gopacket.CaptureInfo
	data    []byte
}

type flowLog struct {
	packets   []pktLog
	last      tick
	timestamp time.Time
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

func uploadToS3(fkey flowKey, flow flowLog, awsRegion, awsS3Bucket string) error {
	buf, err := flow.dumpPcapData()
	if err != nil {
		return errors.Wrap(err, "Fail to dump pcap data")
	}
	logger.WithField("buf_len", len(buf.Bytes())).Trace("Dump pcap data")

	ssn, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
	if err != nil {
		return errors.Wrap(err, "Fail to craete aws session to upload s3")
	}

	key := fmt.Sprintf("pcap/%s/%d_%s_%s.pcap",
		flow.timestamp.Format("2006/01/02/15"),
		flow.timestamp.Unix(),
		strings.Replace(fkey.networkFlow.String(), "->", "_", -1),
		strings.Replace(fkey.transportFlow.String(), "->", "_", -1),
	)

	input := s3.PutObjectInput{
		Bucket: aws.String(awsS3Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	}
	resp, err := s3.New(ssn).PutObject(&input)

	logger.WithFields(logrus.Fields{
		"input":   input,
		"output":  resp,
		"error":   err,
		"flowKey": fkey,
	}).Debug("s3 upload")

	return nil
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

func (x flowKey) String() string {
	return fmt.Sprintf("%s (%s)", x.networkFlow, x.transportFlow)
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

func (x *dataStoreHandler) setS3Bucket(region, s3bucket string) {
	x.awsRegion = region
	x.awsS3Bucket = s3bucket
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
	var extendWaitTime tick = 60

	if !ok {
		flow = &flowLog{
			packets:   []pktLog{},
			last:      x.table.current,
			timestamp: pkt.Metadata().Timestamp,
		}
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

			if x.awsS3Bucket != "" {
				go func() {
					x.wg.Add(1)
					defer x.wg.Done()

					if err := uploadToS3(*fkey, *flow, x.awsRegion, x.awsS3Bucket); err != nil {
						logger.WithError(err).Error("Fail to upload pcpa data on S3 bucket")
					}
				}()
			}

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
	x.wg.Wait()
	return nil
}
