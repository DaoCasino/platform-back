package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"platform-backend/models"
	"time"
)

const statsPath = "/stats"

type AffiliateStatsRepo struct {
	affiliateStatsURL string
}

type GetStatsRequest struct {
	AffiliateID string    `json:"affiliate_id"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
}

func NewAffiliateStatsRepo(affiliateStatsURL string) *AffiliateStatsRepo {
	return &AffiliateStatsRepo{affiliateStatsURL: affiliateStatsURL}
}

func (r *AffiliateStatsRepo) GetStats(
	ctx context.Context, affiliateID string, from time.Time, to time.Time,
) (*models.ReferralStats, error) {
	reqBody, err := json.Marshal(GetStatsRequest{
		AffiliateID: affiliateID,
		From:        from,
		To:          to,
	})
	if err != nil {
		return nil, fmt.Errorf("request body marshal error %w", err)
	}

	resp, err := http.Post(r.affiliateStatsURL+statsPath, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("affiliate-stats get stats request error: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("affiliate-stats respond with error: %s" + resp.Status)
	}

	var stats *models.ReferralStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("affiliate-stats get stats response parsing error: %w", err)
	}

	return stats, nil
}
