package main

import (
	"github.com/jessevdk/go-flags"
	lurker "github.com/m-mizutani/lurker/lib"
	"log"
	"os"
)

type Options struct {
	FileName  string `short:"r" description:"A pcap file" value-name:"FILE"`
	DevName   string `short:"i" description:"Interface name" value-name:"DEV"`
	FluentDst string `short:"f" description:"Destination of fluentd logs" value-name:"HOST:PORT"`
}

func main() {
	var opts Options

	_, ParseErr := flags.ParseArgs(&opts, os.Args)
	if ParseErr != nil {
		os.Exit(1)
	}

	lkr := lurker.Lurker{}
	defer lkr.Close()

	if opts.DevName != "" {
		err := lkr.SetPcapDev(opts.DevName)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	if opts.FileName != "" {
		err := lkr.SetPcapFile(opts.FileName)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	if opts.FluentDst != "" {
		err := lkr.AddFluentdEmitter(opts.FluentDst)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	loopErr := lkr.Loop()
	if loopErr != nil {
		log.Fatal(loopErr)
		os.Exit(1)
	}
}
