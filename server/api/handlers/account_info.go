package handlers

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"strconv"
)

type PlayerInfoResponse struct {
	AccountName      string               `json:"accountName"`
	Email            string               `json:"email"`
	Balance          eos.Asset            `json:"balance"`
	BonusBalances    BonusBalanceResponse `json:"bonusBalances"`
	ActivePermission eos.Authority        `json:"activePermission"`
	OwnerPermission  eos.Authority        `json:"ownerPermission"`
	LinkedCasinos    []*CasinoResponse    `json:"linkedCasinos"`
}

type BonusBalanceResponse map[string]BonusBalance

type BonusBalance struct {
	Balance eos.Asset `json:"balance"`
}

func toPlayerInfoResponse(p *models.PlayerInfo, u *models.User) *PlayerInfoResponse {
	ret := &PlayerInfoResponse{
		AccountName:      u.AccountName,
		Email:            u.Email,
		Balance:          p.Balance,
		BonusBalances:    toBonusBalanceResponse(p.BonusBalances),
		ActivePermission: p.ActivePermission,
		OwnerPermission:  p.OwnerPermission,
		LinkedCasinos:    make([]*CasinoResponse, len(p.LinkedCasinos)),
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

	return toPlayerInfoResponse(player, req.User), nil
}
