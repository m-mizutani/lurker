package tcp_test

import (
	"math/rand"
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/domain/types"
	"github.com/m-mizutani/lurker/pkg/handlers/tcp"
)

func layersToPacket(t *testing.T, f func() []gopacket.SerializableLayer) gopacket.Packet {
	pktLayers := f()

	options := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	buffer := gopacket.NewSerializeBuffer()

	require.NoError(t, gopacket.SerializeLayers(buffer, options, pktLayers...))
	return gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
}

type mockEmitter struct {
	flows []*model.TCPFlow
}

func (x *mockEmitter) Emit(ctx *types.Context, flow *model.TCPFlow) error {
	x.flows = append(x.flows, flow)
	return nil
}

func (x *mockEmitter) Close() error {
	return nil
}

type mockDevice struct {
	wroteData [][]byte
}

func (x *mockDevice) ReadPacket() chan gopacket.Packet {
	panic("must not be called")
}

func (x *mockDevice) WritePacket(pktData []byte) error {
	x.wroteData = append(x.wroteData, pktData)
	return nil
}

func (x *mockDevice) GetDeviceAddrs() ([]net.Addr, error) {
	panic("must not be called")
}

func TestHandleSynPacket(t *testing.T) {
	ctx := types.NewContext()

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

	emitter := &mockEmitter{}
	device := &mockDevice{}
	handler := tcp.New(device, tcp.WithEmitters(interfaces.Emitters{emitter}))

	require.NoError(t, handler.Handle(ctx, synPkt))

	assert.Len(t, device.wroteData, 1)

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

	require.NoError(t, handler.Handle(ctx, ackPkt))
	payload1 := []byte("not ")
	payload2 := []byte("sane")

	require.NoError(t, handler.Handle(ctx, layersToPacket(t, func() []gopacket.SerializableLayer {
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
	})))

	require.NoError(t, handler.Handle(ctx, layersToPacket(t, func() []gopacket.SerializableLayer {
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
	})))

	assert.Empty(t, emitter.flows)
	for i := 0; i < 5; i++ {
		require.NoError(t, handler.Tick(ctx))
	}
	assert.NotEmpty(t, emitter.flows)
}

func TestExcludePort(t *testing.T) {
	baseSeq := rand.Uint32()
	pkt1 := layersToPacket(t, func() []gopacket.SerializableLayer {
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
			DstPort: 5555,
			Ack:     0,
			Seq:     baseSeq,
			SYN:     true,
			Window:  65535,
		}
		require.NoError(t, tcp.SetNetworkLayerForChecksum(ipv4))
		return []gopacket.SerializableLayer{eth, ipv4, tcp}
	})

	pkt2 := layersToPacket(t, func() []gopacket.SerializableLayer {
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
			DstPort: 1111,
			Ack:     0,
			Seq:     baseSeq,
			SYN:     true,
			Window:  65535,
		}
		require.NoError(t, tcp.SetNetworkLayerForChecksum(ipv4))
		return []gopacket.SerializableLayer{eth, ipv4, tcp}
	})

	device := &mockDevice{}
	handler := tcp.New(device, tcp.WithExcludePorts([]int{5555}))

	ctx := types.NewContext()
	handler.Handle(ctx, pkt1)
	assert.Equal(t, 0, len(device.wroteData))

	handler.Handle(ctx, pkt2)
	assert.Equal(t, 1, len(device.wroteData))
}
