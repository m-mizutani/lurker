package lurker

import (
	"testing"
)

func TestLurker(t *testing.T) {
	lkr := New()
	if nil != lkr.SetPcapFile("../test/test_data.pcap") {
		t.Error("can not open file")
	}
}

func TestEmitterQueue(t *testing.T) {
	lurker := New()
	lurker.SetPcapFile("../test/test_data.pcap")
	defer lurker.Close()

	queue, err := NewEmiter("queue")
	if err != nil {
		t.Error("NewEmitter with queue returns nil")
	}

	lurker.AddEmitter(queue)
	lurker.Loop()
}
