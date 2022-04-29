package usecase

import (
	"github.com/m-mizutani/lurker/pkg/domain/handler"
	"github.com/m-mizutani/lurker/pkg/domain/model"
	"github.com/m-mizutani/lurker/pkg/infra"
)

type Usecase struct {
	clients  *infra.Clients
	output   *model.Output
	handlers []handler.Handler
}

func New(clients *infra.Clients, options ...Option) *Usecase {
	uc := &Usecase{
		output:  model.NewOutput(),
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
