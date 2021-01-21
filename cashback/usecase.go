package cashback

import "context"

type UseCase interface {
	CalculateCashback(ctx context.Context, accountName string) (*float64, error)
}
