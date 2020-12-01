package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"platform-backend/referrals/repository/mock"
	"testing"
)

func TestGetOrCreateReferralID(t *testing.T) {
	var (
		repo        = new(mock.ReferralRepoMock)
		refUseCase  = NewReferralsUseCase(repo, true)
		accountName = "somename"
		mockRefID   = "REFblablablablax"
		ctx         = context.Background()
	)

	repo.On("GetReferralID", accountName).Return(mockRefID, nil)

	refID, err := refUseCase.GetOrCreateReferralID(ctx, accountName)
	assert.NoError(t, err)
	assert.Equal(t, mockRefID, refID)

	repo.On("GetReferralID", accountName).Return("", nil)
	repo.On("AddReferralID", accountName, mock2.MatchedBy(func(refID string) bool {
		return true
	})).Return(nil)

	refID, err = refUseCase.GetOrCreateReferralID(ctx, accountName)
	assert.NoError(t, err)
	assert.Regexp(t, "REF[0-9a-zA-Z]{13}", refID)

	refUseCase = NewReferralsUseCase(repo, false)
	refID, err = refUseCase.GetOrCreateReferralID(ctx, accountName)
	assert.NoError(t, err)
	assert.Equal(t, "", refID)
}
