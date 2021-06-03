package utils

import (
	"context"
	"platform-backend/models"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type key int

const (
	keySUID key = iota
	keyUser
	keyRemoteAddr
)

func SetContextSUID(ctx context.Context, suid uuid.UUID) context.Context {
	return context.WithValue(ctx, keySUID, suid)
}

func GetContextSUID(ctx context.Context) (uuid.UUID, bool) {
	suid, ok := ctx.Value(keySUID).(uuid.UUID)
	return suid, ok
}

func SetContextUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, keyUser, user)
}

func GetContextUser(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(keyUser).(*models.User)
	return user, ok
}

func SetContextRemoteAddr(ctx context.Context, ip string) context.Context {
	log.Debug().Msgf("set context remoteAddr: %s", ip)
	return context.WithValue(ctx, keyRemoteAddr, ip)
}

func GetContextRemoteAddr(ctx context.Context) (string, bool) {
	addr, ok := ctx.Value(keyRemoteAddr).(string)
	return addr, ok
}
