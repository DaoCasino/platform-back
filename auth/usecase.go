package auth

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	SignUp(ctx context.Context, user *models.User) (string, string, error)
	SignIn(ctx context.Context, accessToken string) (*models.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}
