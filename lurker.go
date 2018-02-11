package main

import (
	"os"
	"fmt"
	// "log"
	// "github.com/google/gopacket/pcap"
	"github.com/jessevdk/go-flags"
)

func main() {
	var opts struct {
		FileName string `short:"r" description:"A pcap file" value-name:"FILE"`
		DevName string `short:"i" description:"Interface name" value-name:"DEV"`
		FluentDst string `short:"f" description:"Destination of fluentd logs" value-name:"HOST:PORT"`
	}

	args, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		os.Exit(1)
	}

	if opts.FileName != "" {
		fmt.Println(opts.FileName)
	}
	
	fmt.Println(args)
}
