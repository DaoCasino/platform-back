package cashback

import "context"

type Repository interface {
	GetPaidCashback(ctx context.Context, accountName string) (float64, error)
}
