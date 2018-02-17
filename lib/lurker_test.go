package lurker

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLurker(t *testing.T) {
	lkr := New()
	if nil != lkr.SetPcapFile("../test/test_data.pcap") {
		t.Error("can not open file")
	}
}

func NewLurker() (Lurker, *Queue) {
	lurker := New()
	lurker.SetPcapFile("../test/test_data.pcap")
	queue, _ := NewQueue()
	lurker.AddEmitter(queue)
	return lurker, queue
}

func TestEmitterQueueWithoutHandler(t *testing.T) {
	lurker, queue := NewLurker()
	lurker.Loop()

	require.Equal(t, len(queue.Messages), 0,
		"emitter with no handler recieved message(s)")
}

func TestArpSpoofer(t *testing.T) {
	lurker, queue := NewLurker()
	lurker.AddArpSpoofer()
	lurker.Read(2)

	require.Equal(t, 1, len(queue.Messages), "no log by ArpSpoofer")

	m := queue.Messages[0]
	assert.Equal(t, "06:35:8a:6d:7d:37", m["src_hw"], "src_hw is not matched")
	assert.Equal(t, "172.30.1.1", m["src_pr"], "src_pr is not matched")
	assert.Equal(t, "172.30.1.17", m["dst_pr"], "dst_pr is not matched")
}

func TestTcpSpoofer(t *testing.T) {
	lurker, queue := NewLurker()
	require.Equal(t, 0, len(queue.Messages), "not zero")
	lurker.AddTcpSpoofer()
	lurker.Read(20)

	require.Equal(t, 1, len(queue.Messages), "no log by TcpSpoofer")

	m := queue.Messages[0]
	assert.Equal(t, "133.202.248.136", m["src_addr"], "src IP addr is not matched")
}

/*
func TestDataLogger(t *testing.T) {
	lurker, queue := NewLurker()
	lurker.AddDataLogger()
	lurker.Read(40)

	require.Equal(t, 1, len(queue.Messages), "no log by DataLogger")

	m := queue.Messages[0]
	assert.Equal(t, "23.92.220.36", m["src_addr"], "src IP addr is not matched")
}
*/
