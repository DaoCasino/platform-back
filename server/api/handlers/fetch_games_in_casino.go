package handlers

import (
	"context"
	"encoding/json"
	"platform-backend/casino"
	"platform-backend/server/api/interfaces"
)

type FetchGamesInCasinoPayload struct {
	CasinoId uint64  `json:"casinoId"`
}

func ProcessFetchGamesInCasinoRequest(context context.Context, req *interfaces.ApiRequest) (interface{}, *interfaces.HandlerError) {
	var payload FetchGamesInCasinoPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, interfaces.NewHandlerError(interfaces.RequestParseError, err)
	}

	cas, err := req.Repos.Casino.GetCasino(context, payload.CasinoId)
	if err != nil {
		if err == casino.CasinoNotFound {
			return nil, interfaces.NewHandlerError(interfaces.ContentNotFoundError, err)
		}
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	games, err := req.Repos.Casino.GetCasinoGames(context, cas.Contract)
	if err != nil {
		return nil, interfaces.NewHandlerError(interfaces.InternalError, err)
	}

	return games, nil
}
