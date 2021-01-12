package handlers

import (
	"context"
	"math"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"strconv"
	"time"

	"github.com/eoscanada/eos-go"
)

var refStatsFromTime = time.Time{}.Add(time.Second)

type PlayerInfoResponse struct {
	AccountName         string                `json:"accountName"`
	Email               string                `json:"email"`
	Balance             eos.Asset             `json:"balance"`
	BonusBalances       *BonusBalanceResponse `json:"bonusBalances,omitempty"`
	CustomTokenBalances map[string]eos.Asset  `json:"customTokenBalances"`
	ActivePermission    eos.Authority         `json:"activePermission"`
	OwnerPermission     eos.Authority         `json:"ownerPermission"`
	LinkedCasinos       []*CasinoResponse     `json:"linkedCasinos"`
	ReferralID          *string               `json:"referralId,omitempty"`
	ReferralRevenue     *float64              `json:"referralRevenue,omitempty"`
	Referral            *ReferralResponse     `json:"referral,omitempty"`
}

type BonusBalanceResponse map[string]BonusBalance

type BonusBalance struct {
	Balance eos.Asset `json:"balance"`
}

type ReferralResponse struct {
	ID      string  `json:"id"`
	Revenue float64 `json:"revenue"`
}

func toPlayerInfoResponse(
	p *models.PlayerInfo, u *models.User, refID string, refStats *models.ReferralStats,
) *PlayerInfoResponse {
	ret := &PlayerInfoResponse{
		AccountName:         u.AccountName,
		Email:               u.Email,
		Balance:             p.Balance,
		BonusBalances:       toBonusBalanceResponse(p.BonusBalances),
		CustomTokenBalances: p.CustomTokenBalances,
		ActivePermission:    p.ActivePermission,
		OwnerPermission:     p.OwnerPermission,
		LinkedCasinos:       make([]*CasinoResponse, len(p.LinkedCasinos)),
	}
	if refID != "" {
		ret.ReferralID = &refID
		ret.Referral = &ReferralResponse{ID: refID}
	}
	if refStats != nil {
		refStats.ProfitSum = math.Max(0, refStats.ProfitSum)
		ret.ReferralRevenue = &refStats.ProfitSum
		ret.Referral.Revenue = refStats.ProfitSum
	}
	for i, casino := range p.LinkedCasinos {
		ret.LinkedCasinos[i] = toCasinoResponse(casino)
	}

	return ret
}

func toBonusBalanceResponse(bb []*models.BonusBalance) *BonusBalanceResponse {
	if bb == nil {
		return nil
	}

	bbr := make(BonusBalanceResponse)
	for _, b := range bb {
		bbr[strconv.FormatUint(b.CasinoId, 10)] = BonusBalance{Balance: b.Balance}
	}

	return &bbr
}

func ProcessAccountInfo(
	context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	player, err := req.Repos.Contracts.GetPlayerInfo(context, req.User.AccountName)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	refID, err := req.UseCases.Referrals.GetOrCreateReferralID(context, req.User.AccountName)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	refStats, err := req.Repos.AffiliateStats.GetStats(context, refID, refStatsFromTime, time.Now())
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return toPlayerInfoResponse(player, req.User, refID, refStats), nil
}
