package cmd

import (
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/handlers/tcp"
)

func configureExcludePorts(ports []int) ([]tcp.Option, error) {
	for _, port := range ports {
		if port < 0 || 0xffff < port {
			return nil, goerr.New("exclude port must be between 0 to 65536")
		}
	}

	return []tcp.Option{tcp.WithExcludePorts(ports)}, nil
}
