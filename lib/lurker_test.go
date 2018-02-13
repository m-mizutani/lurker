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

	if len(queue.Messages) > 0 {
		t.Error("emitter with no handler recieved message(s)")
	}
}

func TestArpSpoofer(t *testing.T) {
	lurker, queue := NewLurker()
	lurker.AddArpSpoofer()
	lurker.Read(2)

	if len(queue.Messages) != 1 {
		t.Fatal("no log by ArpSpoofer")
	}

	m := queue.Messages[0]
	if m["src_hw"] != "06:35:8a:6d:7d:37" {
		t.Error("src_hw is not matched,", m["src_hw"])
	}

	if m["src_pr"] != "172.30.1.1" {
		t.Error("src_pr is not matched,", m["src_pr"])
	}

	if m["dst_pr"] != "172.30.1.17" {
		t.Error("dst_pr is not matched,", m["dst_pr"])
	}
}
