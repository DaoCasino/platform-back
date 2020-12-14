package usecase

import (
	"context"
	"github.com/stretchr/testify/mock"
	"platform-backend/models"
)

type AuthUseCaseMock struct {
	mock.Mock
}

func (m *AuthUseCaseMock) ResolveUser(ctx context.Context, tmpToken string) (*models.User, error) {
	args := m.Called(tmpToken)

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *AuthUseCaseMock) SignUp(ctx context.Context, user *models.User) (string, string, error) {
	args := m.Called(user)

	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func (m *AuthUseCaseMock) SignIn(ctx context.Context, accessToken string) (*models.User, error) {
	args := m.Called(accessToken)

	return args.Get(0).(*models.User), args.Error(1)
}

func (m *AuthUseCaseMock) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	args := m.Called(refreshToken)

	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func (m *AuthUseCaseMock) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(accessToken)

	return args.Error(0)
}

func (m *AuthUseCaseMock) OptOut(ctx context.Context, accessToken string) error {
	args := m.Called(accessToken)

	return args.Error(0)
}
