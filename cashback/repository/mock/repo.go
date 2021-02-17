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

func (r *CashbackRepoMock) AddUser(ctx context.Context, accountName string) error {
	args := r.Called(accountName)

	return args.Error(0)
}

func (r *CashbackRepoMock) DeleteEthAddress(ctx context.Context, accountName string) error {
	args := r.Called(accountName)

	return args.Error(0)
}

func (r *CashbackRepoMock) SetEthAddress(ctx context.Context, accountName string, ethAddress string) error {
	args := r.Called(accountName, ethAddress)

	return args.Error(0)
}

func (r *CashbackRepoMock) GetEthAddress(ctx context.Context, accountName string) (*string, error) {
	args := r.Called(accountName)

	return args.Get(0).(*string), args.Error(1)
}
