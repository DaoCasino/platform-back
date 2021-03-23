package cashback

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	CashbackInfo(ctx context.Context, accountName string) (*models.CashbackInfo, error)
	SetEthAddress(ctx context.Context, accountName string, ethAddress string) error
	SetStateClaim(ctx context.Context, accountName string) error
	GetCashbacksForClaimed(ctx context.Context) ([]*models.Cashback, error)
	PayCashback(ctx context.Context, accountName string) error
}
