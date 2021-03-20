package cashback

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	CashbackInfo(ctx context.Context, accountName string) (*models.CashbackInfo, error)
	SetEthAddress(ctx context.Context, accountName string, ethAddress string) error
	SetStateAccrued(ctx context.Context, accountName string) error
	SetStateClaim(ctx context.Context, accountName string) error
	GetCashbacksForClaimed(ctx context.Context) ([]*models.Cashback, error)
}
