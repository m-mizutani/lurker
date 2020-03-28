package main

import "github.com/sirupsen/logrus"

func init() {
	logger.SetLevel(logrus.DebugLevel)
}

// Lurker
var NewLurker = newLurker //nolint

func (x *lurker) SetPcapFile(a string) error {
	return x.setPcapFile(a)
}
func (x *lurker) Loop() error {
	return x.loop()
}

// timerTable
var NewTimerTable = newTimerTable //nolint

type Tick tick
type TimerCallback func(Tick) Tick

func (x *timerTable) Add(delay Tick, callback TimerCallback) error {
	return x.add(tick(delay), func(t tick) tick {
		return tick(callback(Tick(t)))
	})
}
func (x *timerTable) Update(count tick) {
	x.update(count)
}
func (x *timerTable) Flush() {
	x.flush()
}
