package main

import "github.com/sirupsen/logrus"

var NewLurker = newLurker

func (x *lurker) SetPcapFile(a string) error {
	return x.setPcapFile(a)
}
func (x *lurker) Loop() error {
	return x.loop()
}

func init() {
	logger.SetLevel(logrus.DebugLevel)
}
