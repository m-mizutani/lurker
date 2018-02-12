package main

import (
	"os"
	"log"
	"github.com/jessevdk/go-flags"
)

func main() {
	var opts Options
	
	_, ParseErr := flags.ParseArgs(&opts, os.Args)
	if ParseErr != nil {
		os.Exit(1)
	}

	lurker := Lurker{}
	defer lurker.Close()
	
	if opts.DevName != "" {
		err := lurker.AddPcapDev(opts.DevName)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}
	
	if opts.FileName != "" {
		err := lurker.AddPcapFile(opts.FileName)
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
