package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type ReferralRepoMock struct {
	mock.Mock
}

func (r *ReferralRepoMock) GetReferralID(ctx context.Context, accountName string) (string, error) {
	args := r.Called(accountName)

	return args.String(0), args.Error(1)
}

func (r *ReferralRepoMock) AddReferralID(ctx context.Context, accountName string, referralID string) error {
	args := r.Called(referralID)

	return args.Error(0)
}

func (r *ReferralRepoMock) GetTotalReferred(ctx context.Context, referralID string) (int, error) {
	args := r.Called(referralID)

	return args.Int(0), args.Error(1)
}
