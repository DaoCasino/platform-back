package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	mock2 "platform-backend/affiliatestats/repository/mock"
	"platform-backend/cashback/repository/mock"
	"testing"
)

func TestCalculateCashback(t *testing.T) {
	var (
		mockCashbackRepo = new(mock.CashbackRepoMock)
		mockAffStatsRepo = new(mock2.AffiliateStatsRepoMock)
		cashbackRatio    = 0.1
		ethToBetRate     = 0.000001
		userGGRs         = map[string]float64{
			"BET": 1234567,
		}
		accountName           = "daosomeuser"
		cashbackUC            = NewCashbackUseCase(mockCashbackRepo, mockAffStatsRepo, cashbackRatio, ethToBetRate, true)
		ctx                   = context.Background()
		expectedCashbackValue = 0.1184567
		expectedCashback      = &expectedCashbackValue
	)

	mockCashbackRepo.On("GetPaidCashback", accountName).Return(0.005, nil)
	mockAffStatsRepo.On("GetUserGGR", accountName).Return(userGGRs, nil)

	cashback, err := cashbackUC.CalculateCashback(ctx, accountName)
	assert.NoError(t, err)
	assert.Equal(t, expectedCashback, cashback)
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
