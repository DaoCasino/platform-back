package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	mock2 "platform-backend/affiliatestats/repository/mock"
	"platform-backend/cashback/repository/mock"
	"platform-backend/models"
	"testing"
)

func TestCashbackInfo(t *testing.T) {

	const (
		cashbackRatio = 0.1
		ethToBetRate  = 0.000001
		paid          = 0.005
		state         = "accrued"
		accountName   = "daosomeuser"
		toPay         = 0.1184567
	)
	var (
		mockCashbackRepo = new(mock.CashbackRepoMock)
		mockAffStatsRepo = new(mock2.AffiliateStatsRepoMock)
		userGGRs         = map[string]float64{
			"BET": 1234567,
		}
		cashbackUC = NewCashbackUseCase(mockCashbackRepo, mockAffStatsRepo, cashbackRatio, ethToBetRate, true)
		ctx        = context.Background()

		row = &models.CashbackRow{
			PaidCashback: paid,
			State:        state,
		}
		expectedCashbackInfo = &models.CashbackInfo{
			ToPay:        toPay,
			Paid:         paid,
			GGR:          userGGRs["BET"],
			Ratio:        cashbackRatio,
			EthToBetRate: ethToBetRate,
			State:        state,
		}
	)

	mockCashbackRepo.On("FetchOne", accountName).Return(row, nil)
	mockAffStatsRepo.On("GetUserGGR", accountName).Return(userGGRs, nil)

	cashback, err := cashbackUC.CashbackInfo(ctx, accountName)
	assert.NoError(t, err)
	assert.Equal(t, expectedCashbackInfo, cashback)
}

func TestSetEthAddress(t *testing.T) {
	const (
		cashbackRatio  = 0.1
		ethToBetRate   = 0.000001
		accountName    = "daosomeuser"
		ethAddr        = "0x323b5d4c32345ced77393b3530b1eed0f346429d"
		invalidEthAddr = "0xXYZb5d4c32345ced77393b3530b1eed0f346429d"
	)
	var (
		mockCashbackRepo = new(mock.CashbackRepoMock)
		mockAffStatsRepo = new(mock2.AffiliateStatsRepoMock)

		cashbackUC = NewCashbackUseCase(mockCashbackRepo, mockAffStatsRepo, cashbackRatio, ethToBetRate, true)
		ctx        = context.Background()
	)

	mockCashbackRepo.On("SetEthAddress", accountName, ethAddr).Return(nil)

	err := cashbackUC.SetEthAddress(ctx, accountName, invalidEthAddr)
	assert.EqualError(t, ErrNonValidEthAddr, err.Error())

	err = cashbackUC.SetEthAddress(ctx, accountName, ethAddr)
	assert.NoError(t, err)
}

func TestSetStateClaim(t *testing.T) {
	const (
		cashbackRatio = 0.1
		ethToBetRate  = 0.000001
		accountName   = "testuser"
	)
	var (
		mockCashbackRepo = new(mock.CashbackRepoMock)
		mockAffStatsRepo = new(mock2.AffiliateStatsRepoMock)

		ctx        = context.Background()
		cashbackUC = NewCashbackUseCase(mockCashbackRepo, mockAffStatsRepo, cashbackRatio, ethToBetRate, true)
	)
	mockCashbackRepo.On("SetStateClaim", accountName).Return(nil)
	err := cashbackUC.SetStateClaim(ctx, accountName)
	assert.NoError(t, err)
}
