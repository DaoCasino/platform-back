package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	"platform-backend/cashback/repository/mock"
	"testing"
)

func TestCalculateCashback(t *testing.T) {
	var (
		mockRepo      = new(mock.CashbackRepoMock)
		cashbackRatio = 0.1
		ethToBetRate  = 0.000001
		active        = true
		userGGRs      = map[string]float64{
			"BET": 1234567,
		}
		accountName           = "daosomeuser"
		cashbackUC            = NewCashbackUseCase(mockRepo, cashbackRatio, ethToBetRate, active)
		ctx                   = context.Background()
		expectedCashbackValue = 0.1184567
		expectedCashback      = &expectedCashbackValue
	)

	mockRepo.On("GetPaidCashback", accountName).Return(0.005, nil)

	cashback, err := cashbackUC.CalculateCashback(ctx, accountName, userGGRs)
	assert.NoError(t, err)
	assert.Equal(t, expectedCashback, cashback)
}
