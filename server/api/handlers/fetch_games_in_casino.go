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

func ProcessFetchGamesInCasinoRequest(context context.Context, req *interfaces.ApiRequest) (*interfaces.WsResponse, error) {
	var payload FetchGamesInCasinoPayload
	if err := json.Unmarshal(req.Data.Payload, &payload); err != nil {
		return nil, err
	}

	cas, err := req.Repos.Casino.GetCasino(context, payload.CasinoId)
	if err != nil {
		if err == casino.CasinoNotFound {
			return &interfaces.WsResponse{
				Type:   "response",
				Id:     req.Data.Id,
				Status: "error",
				Payload: interfaces.WsError{
					Code:    4003,
					Message: err.Error(),
				},
			}, nil
		}
		return &interfaces.WsResponse{
			Type:   "response",
			Id:     req.Data.Id,
			Status: "error",
			Payload: interfaces.WsError{
				Code:    5000,
				Message: "Casino fetch error: " + err.Error(),
			},
		}, nil
	}

	games, err := req.Repos.Casino.GetCasinoGames(context, cas.Contract)
	if err != nil {
		return &interfaces.WsResponse{
			Type:   "response",
			Id:     req.Data.Id,
			Status: "error",
			Payload: interfaces.WsError{
				Code:    5000,
				Message: "Casino games fetch error: " + err.Error(),
			},
		}, nil
	}

	return &interfaces.WsResponse{
		Type:    "response",
		Id:      req.Data.Id,
		Status:  "ok",
		Payload: games,
	}, nil
}
