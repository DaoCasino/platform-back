package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/casino"
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

	cas, err := req.Repos.Casino.GetCasino(context, payload.CasinoId)
	if err != nil {
		if err == casino.CasinoNotFound {
			return nil, ws_interface.NewHandlerError(ws_interface.ContentNotFoundError, err)
		}
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	games, err := req.Repos.Casino.GetCasinoGames(context, cas.Contract)
	if err != nil {
		return nil, ws_interface.NewHandlerError(ws_interface.InternalError, err)
	}

	return games, nil
}
