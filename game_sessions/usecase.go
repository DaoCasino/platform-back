package gamesessions

import (
	"context"
	"platform-backend/models"
	"time"
)

type UseCase interface {
	CleanExpiredSessions(
		ctx context.Context,
		maxLastUpdate time.Duration,
	) error

	NewSession(
		ctx context.Context,
		casino *models.Casino,
		game *models.Game,
		user *models.User,
		deposit string,
		actionType uint16,
		actionParams []uint64,
	) (*models.GameSession, error)

	GameAction(
		ctx context.Context,
		sessionId uint64,
		actionType uint16,
		actionParams []uint64,
	) error

	GameActionWithDeposit(
		ctx context.Context,
		sessionId uint64,
		actionType uint16,
		actionParams []uint64,
		deposit string,
	) error
}
