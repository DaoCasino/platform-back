package handlers

import (
	"context"
	"github.com/eoscanada/eos-go"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
)

type PlayerInfoResponse struct {
	Balance          eos.Asset         `json:"balance"`
	ActivePermission eos.Authority     `json:"activePermission"`
	OwnerPermission  eos.Authority     `json:"ownerPermission"`
	LinkedCasinos    []*CasinoResponse `json:"linkedCasinos"`
}

func toPlayerInfoResponse(p *models.PlayerInfo) *PlayerInfoResponse {
	ret := &PlayerInfoResponse{
		Balance:          p.Balance,
		ActivePermission: p.ActivePermission,
		OwnerPermission:  p.OwnerPermission,
		LinkedCasinos:    make([]*CasinoResponse, len(p.LinkedCasinos)),
	}
	for i, casino := range p.LinkedCasinos {
		ret.LinkedCasinos[i] = toCasinoResponse(casino)
	}

	return ret
}

func ProcessAccountInfo(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	player, err := req.Repos.Contracts.GetPlayerInfo(context, req.User.AccountName)

	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return toPlayerInfoResponse(player), nil
}
