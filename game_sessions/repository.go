package game_sessions

import (
	"context"
	"platform-backend/models"
)

type GameSessionRepository interface {
	HasGameSession(ctx context.Context, id uint64) (bool, error)
	GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error)
	AddGameSession(ctx context.Context, ses *models.GameSession) error
	DeleteGameSession(ctx context.Context, id uint64) error

	GetGameSessionUpdates(ctx context.Context, id uint64) ([]*models.GameSessionUpdate, error)
	AddGameSessionUpdate(ctx context.Context, upd *models.GameSessionUpdate) error
	DeleteGameSessionUpdates(ctx context.Context, sesId uint64) error
}
