package auth

import (
	"context"
	"platform-backend/models"
)

type UseCase interface {
	ResolveUser(ctx context.Context, tmpToken string) (*models.User, error)
	SignUp(ctx context.Context, user *models.User, casinoName string) (string, string, error) // returns: refreshToken, accessToken, error
	SignIn(ctx context.Context, accessToken string) (*models.User, error)
	Logout(ctx context.Context, accessToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error) // returns: refreshToken, accessToken, error
	OptOut(ctx context.Context, accessToken string) error
	AccountNameFromToken(ctx context.Context, accessToken string) (string, error)
	SignInTestAccount(ctx context.Context, accountName string, hash string) (*models.User, error)
}
