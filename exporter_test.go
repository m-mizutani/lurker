package main

import "github.com/sirupsen/logrus"

func init() {
	logger.SetLevel(logrus.DebugLevel)
}

// Lurker
var NewLurker = newLurker

func (x *lurker) SetPcapFile(a string) error {
	return x.setPcapFile(a)
}
func (x *lurker) Loop() error {
	return x.loop()
}

// timerTable
var NewTimerTable = newTimerTable

type Tick tick
type TimerCallback func(Tick) bool

func (x *timerTable) Add(delay Tick, callback TimerCallback) error {
	return x.add(tick(delay), func(t tick) bool {
		return callback(Tick(t))
	})
}
func (x *timerTable) Update(count tick) {
	x.update(count)
}
func (x *timerTable) Flush() {
	x.flush()
}
