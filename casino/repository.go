package casino

import (
	"context"
	"platform-backend/models"
)

type Repository interface {
	GetCasino(ctx context.Context, casinoId uint64) (*models.Casino, error)
	AllCasinos(ctx context.Context) ([]*models.Casino, error)
}
