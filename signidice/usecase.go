package signidice

import "context"

type UseCase interface {
	PerformSignidice(ctx context.Context, gameName string, digest []byte, bcSessionID uint64) error
}
