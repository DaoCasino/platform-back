package casino

import "context"

type UseCase interface {
	GameAction(ctx context.Context, sessionId uint64, actionType uint16, actionParams []uint32) error
}