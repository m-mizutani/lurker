package usecase

import (
	"github.com/m-mizutani/lurker/pkg/domain/interfaces"
	"github.com/m-mizutani/lurker/pkg/infra"
)

type Usecase struct {
	clients  *infra.Clients
	spouts   *interfaces.Spout
	handlers []interfaces.Handler
}

func New(clients *infra.Clients, options ...Option) *Usecase {
	uc := &Usecase{
		spouts:  interfaces.NewSpout(clients),
		clients: clients,
	}

	for _, opt := range options {
		opt(uc)
	}

	return uc
}

type Option func(uc *Usecase)

func WithHandler(hdlr interfaces.Handler) Option {
	return func(uc *Usecase) {
		uc.handlers = append(uc.handlers, hdlr)
	}
}

func WithSpout(spout *interfaces.Spout) Option {
	return func(uc *Usecase) {
		uc.spouts = spout
	}
}
