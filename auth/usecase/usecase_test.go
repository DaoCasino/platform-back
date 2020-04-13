package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"platform-backend/auth/repository/mock"
	"platform-backend/models"
	smMockRepo "platform-backend/server/session_manager/repository/mock"
	"testing"
)

func TestAuthFlow(t *testing.T) {
	repo := new(mock.UserStorageMock)
	sm := new(smMockRepo.MockRepository)

	uc := NewAuthUseCase(repo, sm, []byte("secret"), 10, 10)

	var (
		accountName = "user"
		email = "user@user.com"
		suid, _ = uuid.NewRandom()

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email: email,
		}
	)

	var tokenNonce int64
	tokenNonce = 0

	// Sign Up (Get auth token)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("CreateUser", user).Return(nil)
	repo.On("AddUser", user).Return(nil)
	repo.On("UpdateTokenNonce", user.AccountName).Return(nil)
	repo.On("GetTokenNonce", user.AccountName).Return(tokenNonce, nil)
	_, accessToken, err := uc.SignUp(ctx, user)
	assert.NoError(t, err)

	// Auth with access token
	ctx = context.WithValue(ctx, "suid", suid)
	repo.On("GetUser", user.AccountName).Return(user, nil)
	sm.On("SetUser", suid, user).Return(nil)
	parsedUser, err := uc.SignIn(ctx, accessToken)
	assert.NoError(t, err)
	assert.Equal(t, user, parsedUser)
}

func TestTokenRefresh(t *testing.T) {
	repo := new(mock.UserStorageMock)
	sm := new(smMockRepo.MockRepository)

	uc := NewAuthUseCase(repo, sm, []byte("secret"), 10, 10)

	var (
		accountName = "user"
		email = "user@user.com"
		suid, _ = uuid.NewRandom()

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email: email,
		}
	)

	var tokenNonce int64
	tokenNonce = 0

	// Sign Up (Get auth tokens)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("CreateUser", user).Return(nil)
	repo.On("AddUser", user).Return(nil)
	repo.On("UpdateTokenNonce", user.AccountName).Return(nil)
	repo.On("GetTokenNonce", user.AccountName).Return(tokenNonce, nil)
	refreshToken, accessToken, err := uc.SignUp(ctx, user)
	assert.NoError(t, err)

	// Refresh tokens with refresh token
	ctx = context.WithValue(ctx, "suid", suid)
	repo.On("GetUser", user.AccountName).Return(user, nil)
	sm.On("SetUser", suid, user).Return(nil)
	repo.On("GetTokenNonce", user.AccountName).Return(tokenNonce + 1, nil)
	refreshToken, accessToken, err = uc.RefreshToken(ctx, refreshToken)
	assert.NoError(t, err)

	// Auth with access token
	ctx = context.WithValue(ctx, "suid", suid)
	repo.On("GetUser", user.AccountName).Return(user, nil)
	sm.On("SetUser", suid, user).Return(nil)
	parsedUser, err := uc.SignIn(ctx, accessToken)
	assert.NoError(t, err)
	assert.Equal(t, user, parsedUser)
}
