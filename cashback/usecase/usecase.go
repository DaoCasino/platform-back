package usecase

import (
	"context"
	"fmt"
	"math"
	"platform-backend/affiliatestats"
	"platform-backend/cashback"
	"platform-backend/models"
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

func (c *CashbackUseCase) toPay(userGGR map[string]float64, paid float64) float64 {
	return math.Max(userGGR[utils.DAOBetAssetSymbol]*c.ethToBetRate*c.cashbackRatio-paid, 0)
}

func (c *CashbackUseCase) CashbackInfo(ctx context.Context, accountName string) (*models.CashbackInfo, error) {
	if !c.active {
		return nil, nil
	}

	userGGR, err := c.affStatsRepo.GetUserGGR(ctx, accountName)
	if err != nil {
		return nil, err
	}

	row, err := c.cashbackRepo.FetchOne(ctx, accountName)
	if err != nil {
		return nil, err
	}

	return &models.CashbackInfo{
		ToPay:        c.toPay(userGGR, row.PaidCashback),
		Paid:         row.PaidCashback,
		GGR:          userGGR[utils.DAOBetAssetSymbol],
		Ratio:        c.cashbackRatio,
		EthToBetRate: c.ethToBetRate,
		State:        row.State,
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

func (c *CashbackUseCase) SetStateClaim(ctx context.Context, accountName string) error {
	if !c.active {
		return nil
	}
	return c.cashbackRepo.SetStateClaim(ctx, accountName)
}

func (c *CashbackUseCase) GetCashbacksForClaimed(ctx context.Context) ([]*models.Cashback, error) {
	if !c.active {
		return nil, nil
	}

	rows, err := c.cashbackRepo.FetchAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Cashback, 0, len(rows))
	for _, row := range rows {
		// TODO: need cache, bad ignore error, add GetUsersGGR method
		userGGR, err := c.affStatsRepo.GetUserGGR(ctx, row.AccountName)
		if err != nil {
			return nil, err
		}
		toPay := c.toPay(userGGR, row.PaidCashback)
		if toPay > 0 {
			result = append(result, &models.Cashback{
				AccountName: row.AccountName,
				EthAddress:  row.EthAddress,
				Cashback:    toPay,
			})
		}
	}

	return result, nil
}

func (c *CashbackUseCase) PayCashback(ctx context.Context, accountName string) error {
	if !c.active {
		return nil
	}

	info, err := c.CashbackInfo(ctx, accountName)
	if err != nil {
		return err
	}

	if info.State == "claim" {
		return c.cashbackRepo.SetStateAccrued(ctx, accountName, info.ToPay)
	}

	return fmt.Errorf("state not claim")
}
