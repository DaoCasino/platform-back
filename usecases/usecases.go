package usecases

import (
	"platform-backend/auth"
	"platform-backend/game_sessions"
	"platform-backend/signidice"
)

type UseCases struct {
	Auth        auth.UseCase
	GameSession gamesessions.UseCase
	Signidice   signidice.UseCase
}

func NewUseCases(
	auth auth.UseCase,
	gameSession gamesessions.UseCase,
	signidice signidice.UseCase,
) *UseCases {
	return &UseCases{Auth: auth, GameSession: gameSession, Signidice: signidice}
}
