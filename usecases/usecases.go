package usecases

import (
	"platform-backend/auth"
	"platform-backend/game_sessions"
	"platform-backend/signidice"
	"platform-backend/subscription"
)

type UseCases struct {
	Auth          auth.UseCase
	GameSession   gamesessions.UseCase
	Signidice     signidice.UseCase
	Subscriptions subscription.UseCase
}

func NewUseCases(
	auth auth.UseCase,
	gameSession gamesessions.UseCase,
	signidice signidice.UseCase,
	subscriptions subscription.UseCase,
) *UseCases {
	return &UseCases{Auth: auth, GameSession: gameSession, Signidice: signidice, Subscriptions: subscriptions}
}
