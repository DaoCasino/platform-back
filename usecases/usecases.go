package usecases

import (
	"platform-backend/auth"
	"platform-backend/casino"
)

type UseCases struct {
	Auth   auth.UseCase
	Casino casino.UseCase
}

func NewUseCases(Auth auth.UseCase, Casino casino.UseCase) *UseCases {
	return &UseCases{
		Auth:   Auth,
		Casino: Casino,
	}
}
