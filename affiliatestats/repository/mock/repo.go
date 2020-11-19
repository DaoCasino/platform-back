package mock

import (
	"context"
	"github.com/stretchr/testify/mock"
	"platform-backend/models"
	"time"
)

type AffiliateStatsRepoMock struct {
	mock.Mock
}

func (r *AffiliateStatsRepoMock) GetStats(
	ctx context.Context, affiliateID string, from time.Time, to time.Time,
) (*models.ReferralStats, error) {
	args := r.Called(affiliateID, from, to)

	return args.Get(0).(*models.ReferralStats), args.Error(1)
}
