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
