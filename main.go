package main

import (
	"os"
	"log"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	FileName string `short:"r" description:"A pcap file" value-name:"FILE"`
	DevName string `short:"i" description:"Interface name" value-name:"DEV"`
	FluentDst string `short:"f" description:"Destination of fluentd logs" value-name:"HOST:PORT"`
}

func main() {
	var opts Options
	
	_, ParseErr := flags.ParseArgs(&opts, os.Args)
	if ParseErr != nil {
		os.Exit(1)
	}

	lurker := Lurker{}
	defer lurker.Close()
	
	if opts.DevName != "" {
		err := lurker.SetPcapDev(opts.DevName)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}
	
	if opts.FileName != "" {
		err := lurker.SetPcapFile(opts.FileName)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	if opts.FluentDst != "" {
		err := lurker.AddFluentdEmitter(opts.FluentDst)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}
	
	loopErr := lurker.Loop()
	if loopErr != nil {
		log.Fatal("No available device and pcap file, -i or -r option is mandatory")
		os.Exit(1)
	}
}
