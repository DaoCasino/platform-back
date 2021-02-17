package cashback

import "context"

type Repository interface {
	GetPaidCashback(ctx context.Context, accountName string) (float64, error)
	AddUser(ctx context.Context, accountName string) error
	DeleteEthAddress(ctx context.Context, accountName string) error
	SetEthAddress(ctx context.Context, accountName string, ethAddress string) error
	GetEthAddress(ctx context.Context, accountName string) (*string, error)
}
