package referrals

import "context"

type UseCase interface {
	GetOrCreateReferralID(ctx context.Context, accountName string) (string, error)
}
