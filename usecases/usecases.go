package usecases

import (
	"platform-backend/auth"
	"platform-backend/game_sessions"
)

type UseCases struct {
	Auth        auth.UseCase
	GameSession gamesessions.UseCase
}

func NewUseCases(auth auth.UseCase, gameSession gamesessions.UseCase) *UseCases {
	return &UseCases{Auth: auth,  GameSession: gameSession}
}
