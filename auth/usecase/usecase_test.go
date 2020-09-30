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

	uc := NewAuthUseCase(
		repo,
		sm,
		[]byte("secret"),
		10,
		10,
		"",
		0,
		"",
	)

	var (
		accountName = "user"
		email       = "user@user.com"
		suid, _     = uuid.NewRandom()
		affiliateID = ""

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email:       email,
		}
	)

	tokenNonce := int64(0)
	nextTokenNonce := int64(1)

	// Sign Up (Get auth token)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("CreateUser", user).Return(nil)
	repo.On("AddUser", user).Return(nil)
	repo.On("IsSessionActive", user.AccountName, tokenNonce).Return(true, nil)
	repo.On("InvalidateSession", user.AccountName).Return(nil)
	repo.On("AddNewSession", user.AccountName).Return(nextTokenNonce, nil)
	_, accessToken, err := uc.SignUp(ctx, user, affiliateID)
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

	uc := NewAuthUseCase(
		repo,
		sm,
		[]byte("secret"),
		10,
		10,
		"",
		0,
		"",
	)

	var (
		accountName = "user"
		email       = "user@user.com"
		suid, _     = uuid.NewRandom()
		affiliateID = ""

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email:       email,
		}
	)

	tokenNonce := int64(0)
	nextTokenNonce := int64(1)

	// Sign Up (Get auth tokens)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("CreateUser", user).Return(nil)
	repo.On("AddUser", user).Return(nil)
	repo.On("IsSessionActive", user.AccountName, tokenNonce).Return(true, nil)
	repo.On("InvalidateSession", user.AccountName, tokenNonce).Return(nil)
	repo.On("AddNewSession", user.AccountName).Return(nextTokenNonce, nil)
	refreshToken, _, err := uc.SignUp(ctx, user, affiliateID)
	assert.NoError(t, err)

	// Refresh tokens with refresh token
	ctx = context.WithValue(ctx, "suid", suid)
	repo.On("GetUser", user.AccountName).Return(user, nil)
	sm.On("SetUser", suid, user).Return(nil)
	repo.On("GetLastTokenNonce", user.AccountName).Return(tokenNonce+1, nil)
	_, accessToken, err := uc.RefreshToken(ctx, refreshToken)
	assert.NoError(t, err)

	// Auth with access token
	ctx = context.WithValue(ctx, "suid", suid)
	repo.On("GetUser", user.AccountName).Return(user, nil)
	sm.On("SetUser", suid, user).Return(nil)
	parsedUser, err := uc.SignIn(ctx, accessToken)
	assert.NoError(t, err)
	assert.Equal(t, user, parsedUser)
}

func TestSignUpWithAffiliate(t *testing.T) {
	repo := new(mock.UserStorageMock)
	sm := new(smMockRepo.MockRepository)

	uc := NewAuthUseCase(
		repo,
		sm,
		[]byte("secret"),
		10,
		10,
		"",
		0,
		"",
	)

	var (
		accountName = "user"
		email       = "user@user.com"
		affiliateID = "someAffiliateID"

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email:       email,
		}
	)

	nextTokenNonce := int64(1)

	// Sign Up (Get auth token)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("AddUserWithAffiliate", user, affiliateID).Return(nil)
	repo.On("AddNewSession", user.AccountName).Return(nextTokenNonce, nil)
	_, _, err := uc.SignUp(ctx, user, affiliateID)
	assert.NoError(t, err)
}
