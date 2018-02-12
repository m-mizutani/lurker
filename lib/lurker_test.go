package lurker

import (
	"testing"
)

func TestLurker(t *testing.T) {
	lkr := Lurker{}
	if nil != lkr.SetPcapFile("../test/test_data.pcap") {
		t.Error("can not open file")
	}	
}
