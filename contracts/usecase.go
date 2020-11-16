package contracts

import "context"

type UseCase interface {
	SendNewPlayerToCasino(ctx context.Context, accountName string, casinoName string) error
}
