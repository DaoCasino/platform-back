package usecase

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type ContractsUseCaseMock struct {
	mock.Mock
}

func (m *ContractsUseCaseMock) SendNewPlayerToCasino(ctx context.Context, accountName string, casinoName string) error {
	args := m.Called(accountName, casinoName)

	return args.Error(0)
}
