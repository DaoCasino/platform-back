package auth

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	ResolveUser(ctx context.Context, tmpToken string) (*models.User, error)
	SignUp(ctx context.Context, user *models.User) (string, string, error) // returns: refreshToken, accessToken, error
	SignIn(ctx context.Context, accessToken string) (*models.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error) // returns: refreshToken, accessToken, error
}
