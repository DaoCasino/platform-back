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
	Cashback            *float64              `json:"cashback,omitempty"`
}

type BonusBalanceResponse map[string]BonusBalance

type BonusBalance struct {
	Balance eos.Asset `json:"balance"`
}

type ReferralResponse struct {
	ID            string               `json:"id"`
	Revenue       float64              `json:"revenue"`
	TotalReferred int                  `json:"totalReferred"`
	RevenueToken  map[string]eos.Asset `json:"revenueToken"`
}

func toPlayerInfoResponse(
	p *models.PlayerInfo,
	u *models.User,
	ref *models.Referral,
	refStats *models.ReferralStats,
	cashback *float64,
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
		Cashback:            cashback,
	}
	if ref != nil {
		ret.ReferralID = &ref.ID
		ret.Referral = &ReferralResponse{ID: ref.ID, TotalReferred: ref.TotalReferred}
	}
	if refStats != nil {
		ret.Referral.RevenueToken = make(map[string]eos.Asset)
		refStats.ProfitSum = math.Max(0, refStats.ProfitSum)
		ret.ReferralRevenue = &refStats.ProfitSum
		ret.Referral.Revenue = refStats.ProfitSum
		for key := range refStats.Data {
			ret.Referral.RevenueToken[key] = eos.Asset{
				Amount: eos.Int64(math.Max(0, float64(refStats.Data[key].ProfitSumAsset.Amount))),
				Symbol: refStats.Data[key].ProfitSumAsset.Symbol,
			}
		}
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

	ref, err := req.UseCases.Referrals.GetOrCreateReferral(context, req.User.AccountName)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	var refStats *models.ReferralStats
	if ref != nil {
		refStats, err = req.Repos.AffiliateStats.GetStats(context, ref.ID, refStatsFromTime, time.Now())
		if err != nil {
			return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
		}
	}

	userGGR, err := req.Repos.AffiliateStats.GetUserGGR(context, req.User.AccountName)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}
	cashback, err := req.UseCases.Cashback.CalculateCashback(context, req.User.AccountName, userGGR)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return toPlayerInfoResponse(player, req.User, ref, refStats, cashback), nil
}
