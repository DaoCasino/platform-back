package referrals

import "context"

type Repository interface {
	HasReferralID(ctx context.Context, accountName string) (bool, error)
	GetReferralID(ctx context.Context, accountName string) (string, error)
	AddReferralID(ctx context.Context, accountName string, referralID string) error
}
