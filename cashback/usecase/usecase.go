package usecase

import (
	"context"
	"math"
	"platform-backend/cashback"
	"platform-backend/utils"
)

type CashbackUseCase struct {
	cashbackRepo  cashback.Repository
	cashbackRatio float64
	ethToBetRate  float64
	active        bool
}

func NewCashbackUseCase(
	cashbackRepo cashback.Repository,
	cashbackRatio float64,
	ethToBetRate float64,
	active bool,
) *CashbackUseCase {
	return &CashbackUseCase{
		cashbackRepo:  cashbackRepo,
		cashbackRatio: cashbackRatio,
		ethToBetRate:  ethToBetRate,
		active:        active,
	}
}

func (c *CashbackUseCase) CalculateCashback(
	ctx context.Context,
	accountName string,
	userGGR map[string]float64,
) (*float64, error) {
	if !c.active {
		return nil, nil
	}

	paidCashback, err := c.cashbackRepo.GetPaidCashback(ctx, accountName)
	if err != nil {
		return nil, err
	}
	cb := math.Max(userGGR[utils.DAOBetAssetSymbol]*c.ethToBetRate*c.cashbackRatio-paidCashback, 0)
	return &cb, nil
}
