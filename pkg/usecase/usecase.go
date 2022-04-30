package usecase

import (
	"github.com/m-mizutani/lurker/pkg/infra"
	"github.com/m-mizutani/lurker/pkg/service/handler"
	"github.com/m-mizutani/lurker/pkg/service/spout"
)

type Usecase struct {
	clients  *infra.Clients
	spouts   *spout.Spouts
	handlers []handler.Handler
}

func New(clients *infra.Clients, options ...Option) *Usecase {
	uc := &Usecase{
		spouts:  spout.New(clients),
		clients: clients,
	}

	for _, opt := range options {
		opt(uc)
	}

	return uc
}

type Option func(uc *Usecase)

func WithHandler(hdlr handler.Handler) Option {
	return func(uc *Usecase) {
		uc.handlers = append(uc.handlers, hdlr)
	}
}
