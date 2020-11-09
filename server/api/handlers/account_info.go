package handlers

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"strconv"
)

type PlayerInfoResponse struct {
	AccountName      string                  `json:"accountName"`
	Email            string                  `json:"email"`
	Balance          eos.Asset               `json:"balance"`
	BonusBalances    []*BonusBalanceResponse `json:"bonusBalance"`
	ActivePermission eos.Authority           `json:"activePermission"`
	OwnerPermission  eos.Authority           `json:"ownerPermission"`
	LinkedCasinos    []*CasinoResponse       `json:"linkedCasinos"`
}

type BonusBalanceResponse struct {
	Balance  eos.Asset `json:"balance"`
	CasinoId string    `json:"casinoId"`
}

func toPlayerInfoResponse(p *models.PlayerInfo, u *models.User) *PlayerInfoResponse {
	ret := &PlayerInfoResponse{
		AccountName:      u.AccountName,
		Email:            u.Email,
		Balance:          p.Balance,
		BonusBalances:    make([]*BonusBalanceResponse, len(p.BonusBalances)),
		ActivePermission: p.ActivePermission,
		OwnerPermission:  p.OwnerPermission,
		LinkedCasinos:    make([]*CasinoResponse, len(p.LinkedCasinos)),
	}
	for i, casino := range p.LinkedCasinos {
		ret.LinkedCasinos[i] = toCasinoResponse(casino)
	}
	for i, bonusBalance := range p.BonusBalances {
		ret.BonusBalances[i] = toBonusBalanceResponse(bonusBalance)
	}

	return ret
}

func toBonusBalanceResponse(bb *models.BonusBalance) *BonusBalanceResponse {
	return &BonusBalanceResponse{
		Balance:  bb.Balance,
		CasinoId: strconv.FormatUint(bb.CasinoId, 10),
	}
}

func ProcessAccountInfo(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	player, err := req.Repos.Contracts.GetPlayerInfo(context, req.User.AccountName)

	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return toPlayerInfoResponse(player, req.User), nil
}
