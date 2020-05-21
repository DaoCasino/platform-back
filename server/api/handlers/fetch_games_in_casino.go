package handlers

import (
	"context"
	"encoding/json"
	"github.com/eoscanada/eos-go"
	"platform-backend/contracts"
	"platform-backend/models"
	"platform-backend/server/api/ws_interface"
	"strconv"
)

type FetchGamesInCasinoPayload struct {
	CasinoId eos.Uint64 `json:"casinoId"`
}

type GameParamResponse struct {
	Type  uint16 `json:"type"`
	Value string `json:"value"`
}

type CasinoGameResponse struct {
	Id     string              `json:"gameId"`
	Paused bool                `json:"paused"`
	Params []GameParamResponse `json:"params"`
}

func toCasinoGameResponse(g *models.CasinoGame) *CasinoGameResponse {
	ret := &CasinoGameResponse{
		Id:     strconv.FormatUint(g.Id, 10),
		Paused: g.Paused,
	}

	for _, param := range g.Params {
		ret.Params = append(ret.Params, GameParamResponse{
			Type:  param.Type,
			Value: strconv.FormatUint(param.Value, 10),
		})
	}

	return ret
}

func ProcessFetchGamesInCasinoRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload FetchGamesInCasinoPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	cas, err := req.Repos.Contracts.GetCasino(context, uint64(payload.CasinoId))
	if err != nil {
		if err == contracts.CasinoNotFound {
			return nil, ws_interface.NewHandlerError(ws_interface.CasinoNotFoundError, err)
		}
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	games, err := req.Repos.Contracts.GetCasinoGames(context, cas.Contract)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	var response []*CasinoGameResponse
	for _, game := range games {
		response = append(response, toCasinoGameResponse(game))
	}

	return response, nil
}
