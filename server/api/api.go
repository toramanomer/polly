package api

import "github.com/toramanomer/polly/repository"

type API struct {
	repository *repository.Repository
}

func NewAPI(repository *repository.Repository) *API {
	return &API{
		repository: repository,
	}
}
