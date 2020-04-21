package gamesessions

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	NewSession(ctx context.Context, Casino *models.Casino, Game *models.Game, User *models.User, Deposit string) (*models.GameSession, error)
	GameAction(ctx context.Context, sessionId uint64, actionType uint16, actionParams []uint32) error
	HasGameSession(ctx context.Context, id uint64) (bool, error)
	GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error)
}
