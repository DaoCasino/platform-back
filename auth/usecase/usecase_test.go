package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"platform-backend/auth/repository/mock"
	mock2 "platform-backend/cashback/repository/mock"
	"platform-backend/contracts/usecase"
	"platform-backend/models"
	smMockRepo "platform-backend/server/session_manager/repository/mock"
	"testing"
)

func TestAuthFlow(t *testing.T) {
	repo := new(mock.UserStorageMock)
	cbRepo := new(mock2.CashbackRepoMock)
	sm := new(smMockRepo.MockRepository)
	contractUC := new(usecase.ContractsUseCaseMock)

	uc := NewAuthUseCase(
		repo,
		sm,
		cbRepo,
		contractUC,
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
		affiliateID = "affiliate_1"
		casinoName  = "casinoxxxx"

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email:       email,
			AffiliateID: affiliateID,
		}
	)

	tokenNonce := int64(0)
	nextTokenNonce := int64(1)

	// Sign Up (Get auth token)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("AddUser", user).Return(nil)
	cbRepo.On("AddUser", user.AccountName).Return(nil)
	contractUC.On("SendBonusToNewPlayer", user.AccountName, casinoName).Return(nil)
	repo.On("HasEmail", user.AccountName).Return(true, nil)
	repo.On("IsSessionActive", user.AccountName, tokenNonce).Return(true, nil)
	repo.On("InvalidateSession", user.AccountName).Return(nil)
	repo.On("AddNewSession", user.AccountName).Return(nextTokenNonce, nil)
	_, accessToken, err := uc.SignUp(ctx, user, casinoName)
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
	cbRepo := new(mock2.CashbackRepoMock)
	sm := new(smMockRepo.MockRepository)
	contractUC := new(usecase.ContractsUseCaseMock)

	uc := NewAuthUseCase(
		repo,
		sm,
		cbRepo,
		contractUC,
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
		affiliateID = "affiliate_1"
		casinoName  = "casinoxxxx"

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email:       email,
			AffiliateID: affiliateID,
		}
	)

	tokenNonce := int64(0)
	nextTokenNonce := int64(1)

	// Sign Up (Get auth tokens)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("AddUser", user).Return(nil)
	cbRepo.On("AddUser", user.AccountName).Return(nil)
	contractUC.On("SendBonusToNewPlayer", user.AccountName, casinoName).Return(nil)
	repo.On("HasEmail", user.AccountName).Return(true, nil)
	repo.On("IsSessionActive", user.AccountName, tokenNonce).Return(true, nil)
	repo.On("InvalidateSession", user.AccountName, tokenNonce).Return(nil)
	repo.On("AddNewSession", user.AccountName).Return(nextTokenNonce, nil)
	refreshToken, _, err := uc.SignUp(ctx, user, casinoName)
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

func TestSignUpWithoutAffiliate(t *testing.T) {
	repo := new(mock.UserStorageMock)
	sm := new(smMockRepo.MockRepository)
	cbRepo := new(mock2.CashbackRepoMock)
	contractUC := new(usecase.ContractsUseCaseMock)

	uc := NewAuthUseCase(
		repo,
		sm,
		cbRepo,
		contractUC,
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
		affiliateID = ""
		casinoName  = "casinoxxx"

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email:       email,
			AffiliateID: affiliateID,
		}
	)

	nextTokenNonce := int64(1)

	// Sign Up (Get auth token)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("AddUser", user).Return(nil)
	cbRepo.On("AddUser", user.AccountName).Return(nil)
	contractUC.On("SendBonusToNewPlayer", user.AccountName, casinoName).Return(nil)
	repo.On("HasEmail", user.AccountName).Return(true, nil)
	repo.On("AddNewSession", user.AccountName).Return(nextTokenNonce, nil)
	_, _, err := uc.SignUp(ctx, user, casinoName)
	assert.NoError(t, err)
}

func TestOptOut(t *testing.T) {
	repo := new(mock.UserStorageMock)
	sm := new(smMockRepo.MockRepository)
	cbRepo := new(mock2.CashbackRepoMock)
	contractUC := new(usecase.ContractsUseCaseMock)

	uc := NewAuthUseCase(
		repo,
		sm,
		cbRepo,
		contractUC,
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
		affiliateID = ""
		casinoName  = "casinoxxx"

		ctx = context.Background()

		user = &models.User{
			AccountName: accountName,
			Email:       email,
			AffiliateID: affiliateID,
		}
	)

	tokenNonce := int64(0)
	nextTokenNonce := int64(1)
	nextNextTokenNonce := int64(2)

	// Sign Up (Get auth token)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("AddUser", user).Return(nil)
	cbRepo.On("AddUser", user.AccountName).Return(nil)
	contractUC.On("SendBonusToNewPlayer", user.AccountName, casinoName).Return(nil)
	repo.On("HasEmail", user.AccountName).Return(true, nil)
	repo.On("AddNewSession", user.AccountName).Return(nextTokenNonce, nil)
	_, accessToken, err := uc.SignUp(ctx, user, casinoName)
	assert.NoError(t, err)

	repo.On("IsSessionActive", user.AccountName, tokenNonce).Return(true, nil)
	repo.On("DeleteEmail", user.AccountName).Return(nil)
	cbRepo.On("DeleteEthAddress", user.AccountName).Return(nil)
	err = uc.OptOut(ctx, accessToken)
	assert.NoError(t, err)

	repo.On("HasUser", user.AccountName).Return(true, nil)
	repo.On("HasEmail", user.AccountName).Return(false, nil)
	repo.On("AddEmail", user).Return(nil)
	repo.On("AddNewSession", user.AccountName).Return(nextNextTokenNonce, nil)
	_, _, err = uc.SignUp(ctx, user, casinoName)
	assert.NoError(t, err)
}
