package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type CashbackRepoMock struct {
	mock.Mock
}

func (r *CashbackRepoMock) GetPaidCashback(ctx context.Context, accountName string) (float64, error) {
	args := r.Called(accountName)

	return args.Get(0).(float64), args.Error(1)
}
