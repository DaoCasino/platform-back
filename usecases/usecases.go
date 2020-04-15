package usecases

import (
	"platform-backend/auth"
	"platform-backend/casino"
	"platform-backend/player"
)

type UseCases struct {
	Auth   auth.UseCase
	Casino casino.UseCase
	Player player.UseCase
}

func NewUseCases(Auth auth.UseCase, Casino casino.UseCase, Player player.UseCase) *UseCases {
	return &UseCases{
		Auth:   Auth,
		Casino: Casino,
		Player: Player,
	}
}
