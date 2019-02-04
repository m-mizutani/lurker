package main

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

type options struct {
	FileName    string `short:"r" description:"A pcap file" value-name:"FILE"`
	DevName     string `short:"i" description:"Interface name" value-name:"DEV"`
	Target      string `short:"t" description:"Target Address" value-name:"IPADDR"`
	AwsRegion   string `long:"aws-region"`
	AwsS3Bucket string `long:"aws-s3-bucket"`
}

func main() {
	logger.SetLevel(logrus.DebugLevel)

	var opts options

	_, ParseErr := flags.ParseArgs(&opts, os.Args)
	if ParseErr != nil {
		os.Exit(1)
	}

	lkr := newLurker()
	defer lkr.close()

	if opts.DevName != "" {
		err := lkr.setPcapDev(opts.DevName)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	if opts.FileName != "" {
		err := lkr.setPcapFile(opts.FileName)
		if err != nil {
			logger.Fatal(err)
		}
	}

	if opts.Target != "" {
		lkr.targetAddrs = []string{opts.Target}
	}

	if opts.AwsRegion != "" && opts.AwsS3Bucket != "" {
		lkr.setS3Bucket(opts.AwsRegion, opts.AwsS3Bucket)
	}

	if err := lkr.loop(); err != nil {
		logger.Fatal(err)
	}
}
