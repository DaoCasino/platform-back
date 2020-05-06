package gamesessions

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	NewSession(
		ctx context.Context,
		casino *models.Casino,
		game *models.Game,
		user *models.User,
		deposit string,
	) (*models.GameSession, error)

	GameAction(
		ctx context.Context,
		sessionId uint64,
		actionType uint16,
		actionParams []uint64,
	) error
}
