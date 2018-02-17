package lurker

import (
	"github.com/stretchr/testify/assert"
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
	queue := Queue{}
	lurker.AddEmitter(&queue)
	return lurker, &queue
}

func TestEmitterQueueWithoutHandler(t *testing.T) {
	lurker, queue := NewLurker()
	lurker.Loop()

	assert.Equal(t, len(queue.Messages), 0,
		"emitter with no handler recieved message(s)")
}

func TestArpSpoofer(t *testing.T) {
	lurker, queue := NewLurker()
	lurker.AddArpSpoofer()
	lurker.Read(2)

	if len(queue.Messages) != 1 {
		t.Fatal("no log by ArpSpoofer")
	}

	m := queue.Messages[0]
	assert.Equal(t, "06:35:8a:6d:7d:37", m["src_hw"], "src_hw is not matched")
	assert.Equal(t, "172.30.1.1", m["src_pr"], "src_pr is not matched")
	assert.Equal(t, "172.30.1.17", m["dst_pr"], "dst_pr is not matched")
}
