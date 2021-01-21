package cashback

import "context"

type UseCase interface {
	CalculateCashback(ctx context.Context, accountName string, userGGR map[string]float64) (*float64, error)
}
