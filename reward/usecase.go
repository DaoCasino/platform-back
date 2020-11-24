package reward

import "context"

type UseCase interface {
	RewardGameDevs(ctx context.Context) error
}
