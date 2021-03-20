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
	var (
		mockCashbackRepo = new(mock.CashbackRepoMock)
		mockAffStatsRepo = new(mock2.AffiliateStatsRepoMock)
		cashbackRatio    = 0.1
		ethToBetRate     = 0.000001
		userGGRs         = map[string]float64{
			"BET": 1234567,
		}
		paid                 = 0.005
		accountName          = "daosomeuser"
		cashbackUC           = NewCashbackUseCase(mockCashbackRepo, mockAffStatsRepo, cashbackRatio, ethToBetRate, true)
		ctx                  = context.Background()
		toPay                = 0.1184567
		expectedCashbackInfo = &models.CashbackInfo{
			ToPay:        toPay,
			Paid:         paid,
			GGR:          userGGRs["BET"],
			Ratio:        cashbackRatio,
			EthToBetRate: ethToBetRate,
		}
	)

	mockCashbackRepo.On("GetPaidCashback", accountName).Return(paid, nil)
	mockAffStatsRepo.On("GetUserGGR", accountName).Return(userGGRs, nil)

	cashback, err := cashbackUC.CashbackInfo(ctx, accountName)
	assert.NoError(t, err)
	assert.Equal(t, expectedCashbackInfo, cashback)
}

func TestSetEthAddress(t *testing.T) {
	var (
		mockCashbackRepo = new(mock.CashbackRepoMock)
		mockAffStatsRepo = new(mock2.AffiliateStatsRepoMock)
		cashbackRatio    = 0.1
		ethToBetRate     = 0.000001
		accountName      = "daosomeuser"
		cashbackUC       = NewCashbackUseCase(mockCashbackRepo, mockAffStatsRepo, cashbackRatio, ethToBetRate, true)
		ctx              = context.Background()
		ethAddr          = "0x323b5d4c32345ced77393b3530b1eed0f346429d"
		invalidEthAddr   = "0xXYZb5d4c32345ced77393b3530b1eed0f346429d"
	)

	mockCashbackRepo.On("SetEthAddress", accountName, ethAddr).Return(nil)

	err := cashbackUC.SetEthAddress(ctx, accountName, invalidEthAddr)
	assert.EqualError(t, ErrNonValidEthAddr, err.Error())

	err = cashbackUC.SetEthAddress(ctx, accountName, ethAddr)
	assert.NoError(t, err)
}

func TestSetStateClaim(t *testing.T) {
	var (
		mockCashbackRepo = new(mock.CashbackRepoMock)
		mockAffStatsRepo = new(mock2.AffiliateStatsRepoMock)
		cashbackRatio    = 0.1
		ethToBetRate     = 0.000001
		accountName      = "testuser"
		ctx              = context.Background()
		cashbackUC       = NewCashbackUseCase(mockCashbackRepo, mockAffStatsRepo, cashbackRatio, ethToBetRate, true)
	)
	mockCashbackRepo.On("SetStateClaim", accountName).Return(nil)
	err := cashbackUC.SetStateClaim(ctx, accountName)
	assert.NoError(t, err)
}

func TestSetStateAccrued(t *testing.T) {
	var (
		mockCashbackRepo = new(mock.CashbackRepoMock)
		mockAffStatsRepo = new(mock2.AffiliateStatsRepoMock)
		cashbackRatio    = 0.1
		ethToBetRate     = 0.000001
		accountName      = "testuser"
		ctx              = context.Background()
		cashbackUC       = NewCashbackUseCase(mockCashbackRepo, mockAffStatsRepo, cashbackRatio, ethToBetRate, true)
	)
	mockCashbackRepo.On("SetStateAccrued", accountName).Return(nil)
	err := cashbackUC.SetStateAccrued(ctx, accountName)
	assert.NoError(t, err)
}
