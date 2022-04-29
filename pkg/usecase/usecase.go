package usecase

import "github.com/m-mizutani/lurker/pkg/infra"

type Usecase struct {
	clients *infra.Clients
}

func New(clients *infra.Clients, options ...Option) *Usecase {
	return &Usecase{
		clients: clients,
	}
}

type Option func(uc *Usecase)
