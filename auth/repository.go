package auth

import (
	"context"
	"platform-backend/models"
)

type UserRepository interface {
	HasUser(ctx context.Context, accountName string) (bool, error)
	GetUser(ctx context.Context, accountName string) (*models.User, error)
	AddUser(ctx context.Context, user *models.User) error
	IsSessionActive(ctx context.Context, accountName string, nonce int64) (bool, error)
	InvalidateSession(ctx context.Context, accountName string, nonce int64) error
	AddNewSession(ctx context.Context, accountName string) (int64, error)
	InvalidateOldSessions(ctx context.Context) error
	DeleteEmail(ctx context.Context, accountName string) error
	HasEmail(ctx context.Context, accountName string) (bool, error)
	AddEmail(ctx context.Context, user *models.User) error
	GetTestAccountSalt(ctx context.Context) uint64
	UpdateTestAccountSalt(ctx context.Context)
}
