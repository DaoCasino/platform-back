package gamesessions

import (
	"context"
	"platform-backend/models"
)

type Repository interface {
	HasGameSession(ctx context.Context, id uint64) (bool, error)
	GetGameSession(ctx context.Context, id uint64) (*models.GameSession, error)
	GetSessionByBlockChainID(ctx context.Context, bcID uint64) (*models.GameSession, error)
	UpdateSessionState(ctx context.Context, id uint64, newState uint16) error
	AddGameSession(ctx context.Context, ses *models.GameSession) error
	GetAllGameSessions(ctx context.Context) ([]*models.GameSession, error)
	DeleteGameSession(ctx context.Context, id uint64) error

	GetGameSessionUpdates(ctx context.Context, id uint64) ([]*models.GameSessionUpdate, error)
	AddGameSessionUpdate(ctx context.Context, upd *models.GameSessionUpdate) error
	DeleteGameSessionUpdates(ctx context.Context, sesId uint64) error
}
