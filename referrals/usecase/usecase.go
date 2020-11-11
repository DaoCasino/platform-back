package usecase

import (
	"context"
	"github.com/rs/zerolog/log"
	"platform-backend/referrals"
)

const ReferralIDLen = 13

type ReferralsUseCase struct {
	repo referrals.Repository
}

func NewReferralsUseCase(repo referrals.Repository) *ReferralsUseCase {
	return &ReferralsUseCase{repo: repo}
}

func (r *ReferralsUseCase) GetOrCreateReferralID(ctx context.Context, accountName string) (string, error) {
	hasRefID, err := r.repo.HasReferralID(ctx, accountName)
	if err != nil {
		log.Debug().Msgf("Referral ID exist check error: %s", err.Error())
		return "", err
	}

	if !hasRefID {
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

	refID, err := r.repo.GetReferralID(ctx, accountName)
	if err != nil {
		log.Debug().Msgf("Referral ID get error: %s", err.Error())
		return "", err
	}

	return refID, nil
}
