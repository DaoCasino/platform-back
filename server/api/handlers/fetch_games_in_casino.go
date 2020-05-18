package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/contracts"
	"platform-backend/server/api/ws_interface"
)

type FetchGamesInCasinoPayload struct {
	CasinoId uint64 `json:"casinoId"`
}

func ProcessFetchGamesInCasinoRequest(context context.Context, req *ws_interface.ApiRequest) (interface{}, *ws_interface.HandlerError) {
	var payload FetchGamesInCasinoPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.RequestParseError, err)
	}

	cas, err := req.Repos.Contracts.GetCasino(context, payload.CasinoId)
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

	return games, nil
}
