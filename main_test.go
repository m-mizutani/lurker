package main_test

import (
	"testing"

	lurker "github.com/m-mizutani/lurker"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	lkr := lurker.NewLurker()
	err := lkr.SetPcapFile("test/test_data.pcap")
	require.NoError(t, err)
	err = lkr.Loop()
	require.NoError(t, err)
}
