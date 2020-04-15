package auth

import (
	"context"
	"platform-backend/models"
)

type UserRepository interface {
	HasUser(ctx context.Context, accountName string) (bool, error)
	GetUser(ctx context.Context, accountName string) (*models.User, error)
	AddUser(ctx context.Context, user *models.User) error
	UpdateTokenNonce(ctx context.Context, accountName string) error
	GetTokenNonce(ctx context.Context, accountName string) (int64, error)
}
