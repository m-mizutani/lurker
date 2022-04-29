package main

import (
	"os"

	"github.com/m-mizutani/lurker/pkg/controller/cmd"
)

func main() {
	if err := cmd.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
