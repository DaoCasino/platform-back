package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
	"platform-backend/models"
)

type CashbackRepoMock struct {
	mock.Mock
}

func (r *CashbackRepoMock) GetPaidCashback(_ context.Context, accountName string) (float64, error) {
	args := r.Called(accountName)

	return args.Get(0).(float64), args.Error(1)
}

func (r *CashbackRepoMock) AddUser(_ context.Context, accountName string) error {
	args := r.Called(accountName)

	return args.Error(0)
}

func (r *CashbackRepoMock) DeleteEthAddress(_ context.Context, accountName string) error {
	args := r.Called(accountName)

	return args.Error(0)
}

func (r *CashbackRepoMock) SetEthAddress(_ context.Context, accountName string, ethAddress string) error {
	args := r.Called(accountName, ethAddress)

	return args.Error(0)
}

func (r *CashbackRepoMock) GetEthAddress(_ context.Context, accountName string) (*string, error) {
	args := r.Called(accountName)

	return args.Get(0).(*string), args.Error(1)
}

func (r *CashbackRepoMock) SetStateClaim(_ context.Context, accountName string) error {
	args := r.Called(accountName)
	return args.Error(0)
}

func (r *CashbackRepoMock) SetStateAccrued(_ context.Context, accountName string) error {
	args := r.Called(accountName)
	return args.Error(0)
}

func (r *CashbackRepoMock) FetchAll(_ context.Context) ([]*models.CashbackRow, error) {
	args := r.Called()
	return args.Get(0).([]*models.CashbackRow), args.Error(1)
}
