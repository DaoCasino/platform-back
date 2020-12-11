package referrals

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	GetOrCreateReferral(ctx context.Context, accountName string) (*models.Referral, error)
}
