package referrals

import "context"

type Repository interface {
	GetReferralID(ctx context.Context, accountName string) (string, error)
	AddReferralID(ctx context.Context, accountName string, referralID string) error
	GetTotalReferred(ctx context.Context, referralID string) (int, error)
}
