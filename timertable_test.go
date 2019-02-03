package main_test

import (
	"testing"

	lurker "github.com/m-mizutani/lurker"
	"github.com/stretchr/testify/assert"
)

func TestTimer(t *testing.T) {
	table := lurker.NewTimerTable(10)
	i := 0
	p := &i
	table.Add(3, func(t lurker.Tick) bool {
		*p = 7
		return true
	})

	table.Add(4, func(t lurker.Tick) bool {
		*p = 9
		return true
	})

	table.Update(1)
	assert.Equal(t, 0, i)
	table.Update(2)
	assert.Equal(t, 7, i)
	table.Update(1)
	assert.Equal(t, 9, i)
}

func TestTimerExtend(t *testing.T) {
	table := lurker.NewTimerTable(10)
	i := 0
	p := &i
	table.Add(3, func(t lurker.Tick) bool {
		*p++
		return false
	})

	table.Update(2)
	assert.Equal(t, 0, i)
	table.Update(1)
	assert.Equal(t, 1, i)
	table.Update(2)
	assert.Equal(t, 1, i)
	table.Update(1)
	assert.Equal(t, 2, i)
}
