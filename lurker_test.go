package main

import (
	"testing"
	lurker "github.com/m-mizutani/lurker/lib"
)

func TestLurker(t *testing.T) {
	lkr := lurker.Lurker{}
	if nil != lkr.SetPcapFile("./test/test_data.pcap") {
		t.Error("can not open file")
	}	
}
