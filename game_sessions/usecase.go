package gamesessions

import (
	"context"
)

type UseCase interface {
	NewSession(ctx context.Context, playerId uint32) error
}
