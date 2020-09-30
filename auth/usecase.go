package auth

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	ResolveUser(ctx context.Context, tmpToken string) (*models.User, error)
	SignUp(ctx context.Context, user *models.User, affiliateID string) (string, string, error) // returns: refreshToken, accessToken, error
	SignIn(ctx context.Context, accessToken string) (*models.User, error)
	Logout(ctx context.Context, accessToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error) // returns: refreshToken, accessToken, error
}
