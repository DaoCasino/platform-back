package affiliatestats

import (
	"context"
	"platform-backend/models"
	"time"
)

type Repository interface {
	GetStats(ctx context.Context, affiliateID string, from time.Time, to time.Time) (*models.ReferralStats, error)
}
