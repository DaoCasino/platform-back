package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"platform-backend/models"
	"time"
)

const statsPath = "/stats"
const userGGRPath = "/stats/user/ggr"

type AffiliateStatsRepo struct {
	affiliateStatsURL string
	active            bool
}

type GetStatsRequest struct {
	AffiliateID string    `json:"affiliate_id"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
}

func NewAffiliateStatsRepo(affiliateStatsURL string, active bool) *AffiliateStatsRepo {
	return &AffiliateStatsRepo{affiliateStatsURL: affiliateStatsURL, active: active}
}

func (r *AffiliateStatsRepo) GetStats(
	ctx context.Context, affiliateID string, from time.Time, to time.Time,
) (*models.ReferralStats, error) {
	if !r.active {
		return nil, nil
	}

	reqBody, err := json.Marshal(GetStatsRequest{
		AffiliateID: affiliateID,
		From:        from,
		To:          to,
	})
	if err != nil {
		log.Debug().Msgf("affiliate get stats error: %s", err.Error())
		return nil, fmt.Errorf("request body marshal error %w", err)
	}

	resp, err := http.Post(r.affiliateStatsURL+statsPath, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Debug().Msgf("affiliate get stats request error: %s", err.Error())
		return nil, fmt.Errorf("affiliate-stats get stats request error: %w", err)
	}

	if resp.StatusCode != 200 {
		log.Debug().Msgf("affiliate get respond with error: %s" + resp.Status)
		return nil, fmt.Errorf("affiliate-stats respond with error: %s" + resp.Status)
	}

	var stats *models.ReferralStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		log.Debug().Msgf("affiliate get stats response parsing error: %s", err.Error())
		return nil, fmt.Errorf("affiliate-stats get stats response parsing error: %w", err)
	}

	return stats, nil
}

func (r *AffiliateStatsRepo) GetUserGGR(ctx context.Context, accountName string) (map[string]float64, error) {
	resp, err := http.Get(r.affiliateStatsURL + userGGRPath + "?account_name=" + accountName)
	if err != nil {
		log.Debug().Msgf("get user ggr request error: %s", err.Error())
		return nil, fmt.Errorf("get user ggr request error: %w", err)
	}

	if resp.StatusCode != 200 {
		log.Debug().Msgf("get user ggr respond with error: %s" + resp.Status)
		return nil, fmt.Errorf("get user ggr respond with error: %s" + resp.Status)
	}

	var userGGR map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&userGGR); err != nil {
		log.Debug().Msgf("get user ggr response parsing error: %s", err.Error())
		return nil, fmt.Errorf("get user ggr response parsing error: %w", err)
	}

	return userGGR, nil
}
