package usecases

import (
	"platform-backend/auth"
	"platform-backend/casino"
	"platform-backend/game_sessions"
)

type UseCases struct {
	Auth        auth.UseCase
	Casino      casino.UseCase
	GameSession gamesessions.UseCase
}

func NewUseCases(auth auth.UseCase, casino casino.UseCase, gameSession gamesessions.UseCase) *UseCases {
	return &UseCases{Auth: auth, Casino: casino, GameSession: gameSession}
}
