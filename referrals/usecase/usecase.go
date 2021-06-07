package usecase

import (
	"context"
	"github.com/rs/zerolog/log"
	"platform-backend/models"
	"platform-backend/referrals"
)

const ReferralIDLen = 13

type ReferralsUseCase struct {
	repo   referrals.Repository
	active bool
}

func NewReferralsUseCase(repo referrals.Repository, active bool) *ReferralsUseCase {
	return &ReferralsUseCase{repo: repo, active: active}
}

func (r *ReferralsUseCase) GetOrCreateReferral(ctx context.Context, accountName string) (*models.Referral, error) {
	if !r.active {
		return nil, nil
	}

	refID, err := r.repo.GetReferralID(ctx, accountName)
	if err != nil {
		log.Debug().Msgf("Referral ID get error: %s", err.Error())
		return nil, err
	}

	if refID == "" {
		refID, err := referrals.GenerateRandomString(ReferralIDLen)
		if err != nil {
			log.Debug().Msgf("Referral ID generate error: %s", err.Error())
			return nil, err
		}

		refID = "REF" + refID

		err = r.repo.AddReferralID(ctx, accountName, refID)
		if err != nil {
			log.Debug().Msgf("Referral ID add error: %s", err.Error())
			return nil, err
		}
	}

	totalReferred, err := r.repo.GetTotalReferred(ctx, refID)
	if err != nil {
		log.Debug().Msgf("Total referred get error: %s", err.Error())
		return nil, err
	}

	return &models.Referral{ID: refID, TotalReferred: totalReferred}, nil
}
