package infra

import (
	"github.com/m-mizutani/lurker/pkg/infra/network"
	packet "github.com/m-mizutani/lurker/pkg/infra/network"
)

type Clients struct {
	dev network.Device
}

func New(options ...Option) *Clients {
	clients := &Clients{}

	for _, opt := range options {
		opt(clients)
	}

	return clients
}

func (x *Clients) Device() packet.Device {
	if x.dev == nil {
		panic("network device is not configured")
	}
	return x.dev
}

type Option func(*Clients)

func WithNetworkDevice(dev packet.Device) Option {
	return func(c *Clients) {
		c.dev = dev
	}
}
