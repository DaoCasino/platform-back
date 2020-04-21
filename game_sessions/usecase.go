package gamesessions

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	NewSession(ctx context.Context, GameId uint64, CasinoID uint64, Deposit string, User *models.User) (*models.GameSession, error)
	HasGameSession(ctx context.Context, id uint64) (bool, error)
	GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error)
}
