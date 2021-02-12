package cashback

import "context"

type UseCase interface {
	CalculateCashback(ctx context.Context, accountName string) (*float64, error)
	SetEthAddress(ctx context.Context, accountName string, ethAddress string) error
}
