package location

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	GetLocationFromIP(ctx context.Context, ip string) (*models.Location, error)
}
