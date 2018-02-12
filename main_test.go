package main

import (
	Lurker "github.com/m-mizutani/lurker/lib"
	"testing"
)

func TestMain(t *testing.T) {
	lurker := Lurker.New()
	lurker.SetPcapFile("./test/test_data.pcap")
	lurker.Loop()
}
