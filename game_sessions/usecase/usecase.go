package usecase

import (
	"context"
	"platform-backend/game_sessions"
	"platform-backend/models"
)

type GameSessionsUseCase struct {
	repo gamesessions.Repository
}

func NewGameSessionsUseCase(repo gamesessions.Repository) *GameSessionsUseCase {
	return &GameSessionsUseCase{
		repo: repo,
	}
}

func (a *GameSessionsUseCase) NewSession(ctx context.Context, playerId uint64) error {
	gameSession := &models.GameSession{
		ID:              0,
		Player:          "",
		CasinoID:        0,
		GameID:          0,
		BlockchainSesID: 0,
		State:           0,
	}
	return a.repo.AddGameSession(ctx, gameSession)
}
