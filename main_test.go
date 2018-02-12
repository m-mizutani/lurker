package main

import (
	"testing"
	Lurker "github.com/m-mizutani/lurker/lib"
)

func TestMain(t *testing.T) {
	lurker := Lurker.New()
	lurker.SetPcapFile("./test/test_data.pcap")
	lurker.Loop()
}
