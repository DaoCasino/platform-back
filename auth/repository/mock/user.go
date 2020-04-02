package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
	"platform-backend/models"
)

type UserStorageMock struct {
	mock.Mock
}

func (s *UserStorageMock) HasUser(ctx context.Context, accountName string) (bool, error) {
	args := s.Called(accountName)

	return args.Bool(0), args.Error(1)
}

func (s *UserStorageMock) GetUser(ctx context.Context, accountName string) (*models.User, error) {
	args := s.Called(accountName)

	return args.Get(0).(*models.User), args.Error(1)
}

func (s *UserStorageMock) AddUser(ctx context.Context, user *models.User) error {
	args := s.Called(user)

	return args.Error(0)
}
