package usecase

import (
	"context"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"platform-backend/models"
	"platform-backend/referrals/repository/mock"
	"testing"
)

func TestGetOrCreateReferral(t *testing.T) {
	var (
		repo              = new(mock.ReferralRepoMock)
		refUseCase        = NewReferralsUseCase(repo, true)
		accountName       = "somename"
		mockRefID         = "REFblablablablax"
		mockTotalReferred = 2
		expectedRef       = &models.Referral{ID: mockRefID, TotalReferred: mockTotalReferred}
		ctx               = context.Background()
	)

	repo.On("GetReferralID", accountName).Return(mockRefID, nil)
	repo.On("GetTotalReferred", mockRefID).Return(mockTotalReferred, nil)

	ref, err := refUseCase.GetOrCreateReferral(ctx, accountName)
	assert.NoError(t, err)
	assert.Equal(t, expectedRef, ref)

	repo.On("GetReferralID", accountName).Return("", nil)
	repo.On("AddReferralID", accountName, mock2.MatchedBy(func(refID string) bool {
		return true
	})).Return(nil)

	ref, err = refUseCase.GetOrCreateReferral(ctx, accountName)
	assert.NoError(t, err)
	assert.Regexp(t, "REF[0-9a-zA-Z]{13}", ref.ID)

	refUseCase = NewReferralsUseCase(repo, false)
	ref, err = refUseCase.GetOrCreateReferral(ctx, accountName)
	assert.NoError(t, err)
	assert.Nil(t, ref)
}
