package main

import "github.com/pkg/errors"

type tick uint64
type timerCallback func(tick) tick

type timerTask struct {
	delay    tick
	callback timerCallback
}

type timerUnit struct {
	tasks []timerTask
}

type timerTable struct {
	current tick
	maxTick tick
	units   []timerUnit
}

func newTimerTable(maxTick tick) timerTable {
	table := timerTable{
		maxTick: maxTick,
		units:   make([]timerUnit, maxTick),
	}

	for idx := range table.units {
		table.units[idx].tasks = []timerTask{}
	}

	return table
}

func (x *timerTable) add(delay tick, callback timerCallback) error {
	if delay >= x.maxTick {
		return errors.New("Given delay is over than maximum tick")
	}
	if delay <= 0 {
		return errors.New("delay must be over 0")
	}

	p := (delay + x.current) % x.maxTick
	task := timerTask{
		delay:    delay,
		callback: callback,
	}
	x.units[p].tasks = append(x.units[p].tasks, task)

	return nil
}

func (x *timerTable) update(count tick) {
	for i := tick(1); i <= count && i <= x.maxTick; i++ {
		now := x.current + i
		p := now % x.maxTick
		for _, task := range x.units[p].tasks {
			if extend := task.callback(now); extend > 0 {
				// extend
				x.add(i+extend, task.callback)
			}
		}
		x.units[p].tasks = []timerTask{}
	}

	x.current = (x.current + count)
}

func (x *timerTable) flush() {
	for i := tick(0); i < x.maxTick; i++ {
		now := x.current + i
		p := now % x.maxTick
		for _, task := range x.units[p].tasks {
			task.callback(now)
		}
		x.units[p].tasks = []timerTask{}
	}
}
