package firehose_test

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/emitters/firehose"
	"github.com/stretchr/testify/require"
)

func TestFirehoseClient(t *testing.T) {
	region, streamName := os.Getenv("LURKER_FIREHOSE_REGION"), os.Getenv("LURKER_FIREHOSE_STREAM_NAME")

	if region == "" || streamName == "" {
		t.Skip("region and streamName are required")
	}

	eth := &layers.Ethernet{
		SrcMAC:       []byte{0x12, 0x12, 0x12, 0x12, 0x12, 0x12},
		DstMAC:       []byte{0x12, 0x12, 0x12, 0x12, 0x12, 0x12},
		EthernetType: layers.EthernetTypeIPv4,
	}
	ipv4 := &layers.IPv4{
		Version:  4,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
		SrcIP:    net.ParseIP("10.1.2.3"),
		DstIP:    net.ParseIP("192.168.1.2"),
	}
	pktLayers := []gopacket.SerializableLayer{eth, ipv4}

	options := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	buffer := gopacket.NewSerializeBuffer()

	require.NoError(t, gopacket.SerializeLayers(buffer, options, pktLayers...))
	pkt := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)

	client := firehose.New(region, streamName)
	flow := model.TCPFlow{
		SrcHost:   pkt.NetworkLayer().NetworkFlow().Src(),
		DstHost:   pkt.NetworkLayer().NetworkFlow().Dst(),
		SrcPort:   54321,
		DstPort:   80,
		CreatedAt: time.Now(),
		RecvData:  []byte("hello"),
		RecvAck:   true,
	}
	require.NoError(t, client.Emit(types.NewContext(), &flow))
}
