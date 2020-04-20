package usecases

import (
	"platform-backend/auth"
)

type UseCases struct {
	Auth   auth.UseCase
}

func NewUseCases(Auth auth.UseCase) *UseCases {
	return &UseCases{
		Auth:   Auth,
	}
}
