package casino

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	GetCasino(ctx context.Context, casinoId uint64) (*models.Casino, error)
	AllCasinos(ctx context.Context) ([]*models.Casino, error)
}
