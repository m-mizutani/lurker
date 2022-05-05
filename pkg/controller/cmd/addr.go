package cmd

import (
	"net"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/lurker/pkg/handlers/tcp"
	"github.com/m-mizutani/lurker/pkg/infra/network"
)

func configureAddrs(cfg *Config, dev network.Device) ([]tcp.Option, error) {
	var resp []tcp.Option

	addrs := cfg.ListenAddrs.Value()
	if len(addrs) == 0 {
		addrs, err := dev.GetDeviceAddrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, goerr.Wrap(err)
			}
			if ip.To4() == nil {
				continue
			}

			resp = append(resp, tcp.WithAllowedNetwork(&net.IPNet{
				IP:   ip,
				Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0xff),
			}))
		}
	} else {
		for _, addr := range addrs {
			_, ipNet, err := net.ParseCIDR(addr)
			if err != nil {
				return nil, goerr.Wrap(err)
			}

			resp = append(resp, tcp.WithAllowedNetwork(ipNet))
		}

	}

	return resp, nil
}
