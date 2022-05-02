package tcp_test

import (
	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/handlers/tcp"
)

func layersToPacket(t *testing.T, f func() []gopacket.SerializableLayer) gopacket.Packet {
	pktLayers := f()

	options := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	buffer := gopacket.NewSerializeBuffer()

	require.NoError(t, gopacket.SerializeLayers(buffer, options, pktLayers...))
	return gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
}

func TestHandleSynPacket(t *testing.T) {
	baseSeq := rand.Uint32()
	synPkt := layersToPacket(t, func() []gopacket.SerializableLayer {
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
		tcp := &layers.TCP{
			SrcPort: 54321,
			DstPort: 80,
			Ack:     0,
			Seq:     baseSeq,
			SYN:     true,
			Window:  65535,
		}
		require.NoError(t, tcp.SetNetworkLayerForChecksum(ipv4))
		return []gopacket.SerializableLayer{eth, ipv4, tcp}
	})

	var calledWritePacket int
	handler := tcp.New()
	var logOutput string
	spouts := &interfaces.Spout{
		Console: func(format string, args ...any) {
			logOutput = fmt.Sprintf(format, args)
		},
		WritePacket: func(b []byte) {
			calledWritePacket++
		},
	}

	require.NoError(t, handler.Handle(synPkt, spouts))

	assert.Equal(t, 1, calledWritePacket)

	ackPkt := layersToPacket(t, func() []gopacket.SerializableLayer {
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
		tcp := &layers.TCP{
			SrcPort: 54321,
			DstPort: 80,
			Ack:     0,
			Seq:     baseSeq + 1,
			SYN:     true,
			ACK:     true,
			Window:  65535,
		}

		require.NoError(t, tcp.SetNetworkLayerForChecksum(ipv4))
		return []gopacket.SerializableLayer{eth, ipv4, tcp}
	})

	require.NoError(t, handler.Handle(ackPkt, spouts))
	payload1 := []byte("not ")
	payload2 := []byte("sane")

	require.NoError(t, handler.Handle(layersToPacket(t, func() []gopacket.SerializableLayer {
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
		tcp := &layers.TCP{
			SrcPort: 54321,
			DstPort: 80,
			Seq:     baseSeq + 1,
			Window:  65535,
		}

		require.NoError(t, tcp.SetNetworkLayerForChecksum(ipv4))
		payload := gopacket.Payload(payload1)
		return []gopacket.SerializableLayer{eth, ipv4, tcp, payload}
	}), spouts))

	require.NoError(t, handler.Handle(layersToPacket(t, func() []gopacket.SerializableLayer {
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
		tcp := &layers.TCP{
			SrcPort: 54321,
			DstPort: 80,
			Seq:     baseSeq + 1 + uint32(len(payload1)),
			Window:  65535,
		}

		require.NoError(t, tcp.SetNetworkLayerForChecksum(ipv4))
		payload := gopacket.Payload(payload2)
		return []gopacket.SerializableLayer{eth, ipv4, tcp, payload}
	}), spouts))

	assert.Empty(t, logOutput)
	for i := 0; i < 5; i++ {
		require.NoError(t, handler.Tick(spouts))
	}
	assert.NotEmpty(t, logOutput)
}
