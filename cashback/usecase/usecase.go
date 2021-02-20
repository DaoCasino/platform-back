package usecase

import (
	"context"
	"fmt"
	"math"
	"platform-backend/affiliatestats"
	"platform-backend/cashback"
	"platform-backend/utils"
)

var (
	ErrNonValidEthAddr = fmt.Errorf("non valid eth address")
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

func (c *CashbackUseCase) CashbackInfo(ctx context.Context, accountName string) (*cashback.Info, error) {
	if !c.active {
		return nil, nil
	}

	userGGR, err := c.affStatsRepo.GetUserGGR(ctx, accountName)
	if err != nil {
		return nil, err
	}

	paid, err := c.cashbackRepo.GetPaidCashback(ctx, accountName)
	if err != nil {
		return nil, err
	}
	toPay := math.Max(userGGR[utils.DAOBetAssetSymbol]*c.ethToBetRate*c.cashbackRatio-paid, 0)

	return &cashback.Info{
		ToPay:        toPay,
		Paid:         paid,
		GGR:          userGGR[utils.DAOBetAssetSymbol],
		Ratio:        c.cashbackRatio,
		EthToBetRate: c.ethToBetRate,
	}, nil
}

func (c *CashbackUseCase) SetEthAddress(ctx context.Context, accountName string, ethAddress string) error {
	if !c.active {
		return nil
	}

	if !utils.IsValidEthAddress(ethAddress) {
		return ErrNonValidEthAddr
	}

	return c.cashbackRepo.SetEthAddress(ctx, accountName, ethAddress)
}
