package usecase

import (
	"context"
	"github.com/rs/zerolog/log"
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

func (r *ReferralsUseCase) GetOrCreateReferralID(ctx context.Context, accountName string) (string, error) {
	if !r.active {
		return "", nil
	}

	refID, err := r.repo.GetReferralID(ctx, accountName)
	if err != nil {
		log.Debug().Msgf("Referral ID get error: %s", err.Error())
		return "", err
	}

	if refID == "" {
		refID, err := referrals.GenerateRandomString(ReferralIDLen)
		if err != nil {
			log.Debug().Msgf("Referral ID generate error: %s", err.Error())
			return "", err
		}

		refID = "REF" + refID

		err = r.repo.AddReferralID(ctx, accountName, refID)
		if err != nil {
			log.Debug().Msgf("Referral ID add error: %s", err.Error())
			return "", err
		}

		return refID, nil
	}

	return refID, nil
}
