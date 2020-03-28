package main_test

import (
	"testing"

	lurker "github.com/m-mizutani/lurker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimer(t *testing.T) {
	table := lurker.NewTimerTable(10)
	i := 0
	p := &i
	require.NoError(t, table.Add(3, func(t lurker.Tick) lurker.Tick {
		*p = 7
		return 0
	}))

	require.NoError(t, table.Add(4, func(t lurker.Tick) lurker.Tick {
		*p = 9
		return 0
	}))

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
	require.NoError(t, table.Add(3, func(t lurker.Tick) lurker.Tick {
		*p++
		return 3
	}))

	table.Update(2)
	assert.Equal(t, 0, i)
	table.Update(1)
	assert.Equal(t, 1, i)
	table.Update(2)
	assert.Equal(t, 1, i)
	table.Update(1)
	assert.Equal(t, 2, i)
}
