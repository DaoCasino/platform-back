package contracts

import "context"

type UseCase interface {
	SendBonusToNewPlayer(ctx context.Context, accountName string, casinoName string) error
}
