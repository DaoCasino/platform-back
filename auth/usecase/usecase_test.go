package usecase

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"platform-backend/auth/repository/mock"
	"platform-backend/models"
	"platform-backend/server/session"
	"testing"
)

func TestAuthFlow(t *testing.T) {
	repo := new(mock.UserStorageMock)
	sm := new (session.ManagerMock)

	uc := NewAuthUseCase(repo, sm, []byte("secret"))

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

	// Sign Up (Get auth token)
	repo.On("HasUser", user.AccountName).Return(false, nil)
	repo.On("CreateUser", user).Return(nil)
	repo.On("AddUser", user).Return(nil)
	token, err := uc.SignUp(ctx, user)
	assert.NoError(t, err)

	// Verify token
	ctx = context.WithValue(ctx, "suid", suid)
	repo.On("GetUser", user.AccountName).Return(user, nil)
	sm.On("AuthUser", suid, user).Return(nil)
	parsedUser, err := uc.SignIn(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, user, parsedUser)
}
