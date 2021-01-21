package usecase

import (
	"context"
	"math"
	"platform-backend/affiliatestats"
	"platform-backend/cashback"
	"platform-backend/utils"
)

type CashbackUseCase struct {
	cashbackRepo  cashback.Repository
	affStatsRepo  affiliatestats.Repository
	cashbackRatio float64
	ethToBetRate  float64
	active        bool
}

func NewCashbackUseCase(
	cashbackRepo cashback.Repository,
	affStatsRepo affiliatestats.Repository,
	cashbackRatio float64,
	ethToBetRate float64,
	active bool,
) *CashbackUseCase {
	return &CashbackUseCase{
		cashbackRepo:  cashbackRepo,
		affStatsRepo:  affStatsRepo,
		cashbackRatio: cashbackRatio,
		ethToBetRate:  ethToBetRate,
		active:        active,
	}
}

func (c *CashbackUseCase) CalculateCashback(
	ctx context.Context,
	accountName string,
) (*float64, error) {
	if !c.active {
		return nil, nil
	}

	userGGR, err := c.affStatsRepo.GetUserGGR(ctx, accountName)
	if err != nil {
		return nil, err
	}

	paidCashback, err := c.cashbackRepo.GetPaidCashback(ctx, accountName)
	if err != nil {
		return nil, err
	}
	cb := math.Max(userGGR[utils.DAOBetAssetSymbol]*c.ethToBetRate*c.cashbackRatio-paidCashback, 0)
	return &cb, nil
}
