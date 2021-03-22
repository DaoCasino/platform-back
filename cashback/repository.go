package cashback

import "context"
import "platform-backend/models"

type Repository interface {
	GetPaidCashback(ctx context.Context, accountName string) (float64, error)
	AddUser(ctx context.Context, accountName string) error
	DeleteEthAddress(ctx context.Context, accountName string) error
	SetEthAddress(ctx context.Context, accountName string, ethAddress string) error
	GetEthAddress(ctx context.Context, accountName string) (*string, error)
	SetStateClaim(ctx context.Context, accountName string) error
	SetStateAccrued(ctx context.Context, accountName string) error
	FetchAll(ctx context.Context) ([]*models.CashbackRow, error)
	FetchOne(ctx context.Context, accountName string) (*models.CashbackRow, error)
}
