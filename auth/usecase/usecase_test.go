package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	"platform-backend/auth/repository/mock"
	"platform-backend/models"
	"testing"
)

func TestAuthFlow(t *testing.T) {
	repo := new(mock.UserStorageMock)

	uc := NewAuthUseCase(repo, []byte("secret"))

	var (
		accountName = "user"
		email = "user@user.com"

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email: email,
		}
	)

	// Sign Up (Get auth token)
	repo.On("CreateUser", user).Return(nil)
	token, err := uc.SignUp(ctx, user)
	assert.NoError(t, err)

	// Verify token
	parsedUser, err := uc.ParseToken(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, user, parsedUser)
}
