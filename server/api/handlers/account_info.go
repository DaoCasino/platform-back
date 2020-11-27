package handlers

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"strconv"
	"time"
)

var refStatsFromTime = time.Time{}.Add(time.Second)

type PlayerInfoResponse struct {
	AccountName      string               `json:"accountName"`
	Email            string               `json:"email"`
	Balance          eos.Asset            `json:"balance"`
	BonusBalances    BonusBalanceResponse `json:"bonusBalances"`
	ActivePermission eos.Authority        `json:"activePermission"`
	OwnerPermission  eos.Authority        `json:"ownerPermission"`
	LinkedCasinos    []*CasinoResponse    `json:"linkedCasinos"`
	ReferralID       string               `json:"referralId"`
	ReferralRevenue  float64              `json:"referralRevenue"`
}

type BonusBalanceResponse map[string]BonusBalance

type BonusBalance struct {
	Balance eos.Asset `json:"balance"`
}

func toPlayerInfoResponse(
	p *models.PlayerInfo, u *models.User, refID string, refStats *models.ReferralStats,
) *PlayerInfoResponse {
	ret := &PlayerInfoResponse{
		AccountName:      u.AccountName,
		Email:            u.Email,
		Balance:          p.Balance,
		BonusBalances:    toBonusBalanceResponse(p.BonusBalances),
		ActivePermission: p.ActivePermission,
		OwnerPermission:  p.OwnerPermission,
		LinkedCasinos:    make([]*CasinoResponse, len(p.LinkedCasinos)),
		ReferralID:       refID,
		ReferralRevenue:  refStats.ProfitSum,
	}
	for i, casino := range p.LinkedCasinos {
		ret.LinkedCasinos[i] = toCasinoResponse(casino)
	}

	return ret
}

func toBonusBalanceResponse(bb []*models.BonusBalance) BonusBalanceResponse {
	bbr := make(BonusBalanceResponse)

	for _, b := range bb {
		bbr[strconv.FormatUint(b.CasinoId, 10)] = BonusBalance{Balance: b.Balance}
	}

	return bbr
}

func ProcessAccountInfo(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
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
