package player

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	GetInfo(ctx context.Context, accountName string) (*models.PlayerInfo, error)
}