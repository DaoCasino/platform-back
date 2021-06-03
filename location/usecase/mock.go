package usecase

import (
	"context"
	"platform-backend/models"

	"github.com/stretchr/testify/mock"
)

type LocationUseCaseMock struct {
	mock.Mock
}

func (m *LocationUseCaseMock) GetLocationFromIP(_ context.Context, ip string) (*models.Location, error) {
	args := m.Called(ip)

	return args.Get(0).(*models.Location), args.Error(1)
}
