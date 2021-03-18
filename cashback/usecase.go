package cashback

import (
	"context"
)

type Info struct {
	ToPay        float64 `json:"toPay"`
	Paid         float64 `json:"paid"`
	GGR          float64 `json:"ggr"`
	Ratio        float64 `json:"ratio"`
	EthToBetRate float64 `json:"ethToBetRate"`
}

type UseCase interface {
	CashbackInfo(ctx context.Context, accountName string) (*Info, error)
	SetEthAddress(ctx context.Context, accountName string, ethAddress string) error
	SetStateAccrued(ctx context.Context, accountName string) error
	SetStateClaim(ctx context.Context, accountName string) error
}
